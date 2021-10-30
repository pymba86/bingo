package service

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/pymba86/bingo/pkg/exchange/batch"
	"github.com/pymba86/bingo/pkg/types"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

var ErrTradeNotFound = errors.New("trade not found")

type QueryTradesOptions struct {
	Exchange types.ExchangeName
	Symbol   string
	LastGID  int64

	// ASC or DESC
	Ordering string
	Limit    int
}

type TradingVolume struct {
	Year        int       `db:"year" json:"year"`
	Month       int       `db:"month" json:"month,omitempty"`
	Day         int       `db:"day" json:"day,omitempty"`
	Time        time.Time `json:"time,omitempty"`
	Exchange    string    `db:"exchange" json:"exchange,omitempty"`
	Symbol      string    `db:"symbol" json:"symbol,omitempty"`
	QuoteVolume float64   `db:"quote_volume" json:"quoteVolume"`
}

type TradingVolumeQueryOptions struct {
	GroupByPeriod string
	SegmentBy     string
}

type TradeService struct {
	DB *sqlx.DB
}

func NewTradeService(db *sqlx.DB) *TradeService {
	return &TradeService{db}
}

func (s *TradeService) Sync(ctx context.Context, exchange types.Exchange, symbol string) error {
	isMargin := false
	isIsolated := false

	if marginExchange, ok := exchange.(types.MarginExchange); ok {
		marginSettings := marginExchange.GetMarginSettings()
		isMargin = marginSettings.IsMargin
		isIsolated = marginSettings.IsIsolatedMargin
		if marginSettings.IsIsolatedMargin {
			symbol = marginSettings.IsolatedMarginSymbol
		}
	}

	records, err := s.QueryLast(exchange.Name(), symbol, isMargin, isIsolated, 50)
	if err != nil {
		return err
	}
	var tradeKeys = map[types.TradeKey]struct{}{}
	var lastTradeID int64 = 1
	if len(records) > 0 {
		for _, record := range records {
			tradeKeys[record.Key()] = struct{}{}
		}

		lastTradeID = records[0].ID
	}

	b := &batch.TradeBatchQuery{Exchange: exchange}
	tradeC, errC := b.Query(ctx, symbol, &types.TradeQueryOptions{
		LastTradeID: lastTradeID,
	})

	for trade := range tradeC {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := <-errC:
			if err != nil {
				return err
			}

		default:
		}

		key := trade.Key()
		if _, exists := tradeKeys[key]; exists {
			continue
		}

		tradeKeys[key] = struct{}{}

		log.Infof("inserting trade: %s %d %s %-4s price: %-13f volume: %-11f %5s %s",
			trade.Exchange,
			trade.ID,
			trade.Symbol,
			trade.Side,
			trade.Price,
			trade.Quantity,
			trade.Liquidity(),
			trade.Time.String())

		if err := s.Insert(trade); err != nil {
			return err
		}
	}

	return <-errC
}

func (s *TradeService) QueryLast(ex types.ExchangeName, symbol string, isMargin, isIsolated bool, limit int) ([]types.Trade, error) {
	log.Debugf("querying last trade exchange = %s AND symbol = %s AND is_margin = %v AND is_isolated = %v", ex, symbol, isMargin, isIsolated)

	sql := "SELECT * FROM trades WHERE exchange = :exchange AND symbol = :symbol AND is_margin = :is_margin AND is_isolated = :is_isolated ORDER BY gid DESC LIMIT :limit"
	rows, err := s.DB.NamedQuery(sql, map[string]interface{}{
		"symbol":      symbol,
		"exchange":    ex,
		"is_margin":   isMargin,
		"is_isolated": isIsolated,
		"limit":       limit,
	})
	if err != nil {
		return nil, errors.Wrap(err, "query last trade error")
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *TradeService) QueryForTradingFeeCurrency(ex types.ExchangeName, symbol string, feeCurrency string) ([]types.Trade, error) {
	sql := "SELECT * FROM trades WHERE exchange = :exchange AND (symbol = :symbol OR fee_currency = :fee_currency) ORDER BY traded_at ASC"
	rows, err := s.DB.NamedQuery(sql, map[string]interface{}{
		"exchange":     ex,
		"symbol":       symbol,
		"fee_currency": feeCurrency,
	})
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *TradeService) Query(options QueryTradesOptions) ([]types.Trade, error) {
	sql := queryTradesSQL(options)

	log.Info(sql)

	args := map[string]interface{}{
		"exchange": options.Exchange,
		"symbol":   options.Symbol,
	}
	rows, err := s.DB.NamedQuery(sql, args)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *TradeService) Load(ctx context.Context, id int64) (*types.Trade, error) {
	var trade types.Trade

	rows, err := s.DB.NamedQuery("SELECT * FROM trades WHERE id = :id", map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(&trade)
		return &trade, err
	}

	return nil, errors.Wrapf(ErrTradeNotFound, "trade id:%d not found", id)
}

func (s *TradeService) Mark(ctx context.Context, id int64, strategyID string) error {
	result, err := s.DB.NamedExecContext(ctx, "UPDATE `trades` SET `strategy` = :strategy WHERE `id` = :id",
		map[string]interface{}{
			"id":       id,
			"strategy": strategyID,
		})
	if err != nil {
		return err
	}

	cnt, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if cnt == 0 {
		return fmt.Errorf("trade id:%d not found", id)
	}

	return nil
}

func (s *TradeService) UpdatePnL(ctx context.Context, id int64, pnl float64) error {
	result, err := s.DB.NamedExecContext(ctx, "UPDATE `trades` SET `pnl` = :pnl WHERE `id` = :id", map[string]interface{}{
		"id":  id,
		"pnl": pnl,
	})
	if err != nil {
		return err
	}

	cnt, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if cnt == 0 {
		return fmt.Errorf("trade id:%d not found", id)
	}

	return nil

}

func queryTradesSQL(options QueryTradesOptions) string {
	ordering := "ASC"
	switch v := strings.ToUpper(options.Ordering); v {
	case "DESC", "ASC":
		ordering = v
	}

	var where []string

	if len(options.Exchange) > 0 {
		where = append(where, `exchange = :exchange`)
	}

	if len(options.Symbol) > 0 {
		where = append(where, `symbol = :symbol`)
	}

	if options.LastGID > 0 {
		switch ordering {
		case "ASC":
			where = append(where, "gid > :gid")
		case "DESC":
			where = append(where, "gid < :gid")
		}
	}

	sql := `SELECT * FROM trades`

	if len(where) > 0 {
		sql += ` WHERE ` + strings.Join(where, " AND ")
	}

	sql += ` ORDER BY gid ` + ordering

	if options.Limit > 0 {
		sql += ` LIMIT ` + strconv.Itoa(options.Limit)
	}

	return sql
}

func (s *TradeService) scanRows(rows *sqlx.Rows) (trades []types.Trade, err error) {
	for rows.Next() {
		var trade types.Trade
		if err := rows.StructScan(&trade); err != nil {
			return trades, err
		}

		trades = append(trades, trade)
	}

	return trades, rows.Err()
}

func (s *TradeService) Insert(trade types.Trade) error {
	_, err := s.DB.NamedExec(`
			INSERT INTO trades (id, exchange, order_id, symbol, price, quantity, quote_quantity, side, is_buyer, is_maker, fee, fee_currency, traded_at, is_margin, is_isolated)
			VALUES (:id, :exchange, :order_id, :symbol, :price, :quantity, :quote_quantity, :side, :is_buyer, :is_maker, :fee, :fee_currency, :traded_at, :is_margin, :is_isolated)`,
		trade)
	return err
}

func (s *TradeService) DeleteAll() error {
	_, err := s.DB.Exec(`DELETE FROM trades`)
	return err
}
