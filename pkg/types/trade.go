package types

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

func init() {
	// make sure we can cast Trade to PlainText
	_ = PlainText(Trade{})
	_ = PlainText(&Trade{})
}

type TradeSlice struct {
	mu     sync.Mutex
	Trades []Trade
}

func (s *TradeSlice) Copy() []Trade {
	s.mu.Lock()
	slice := make([]Trade, len(s.Trades), len(s.Trades))
	copy(slice, s.Trades)
	s.mu.Unlock()

	return slice
}

func (s *TradeSlice) Reverse() {
	slice := s.Trades
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func (s *TradeSlice) Append(t Trade) {
	s.mu.Lock()
	s.Trades = append(s.Trades, t)
	s.mu.Unlock()
}

type Trade struct {
	// GID is the global ID
	GID int64 `json:"gid" db:"gid"`

	// ID is the source trade ID
	ID            int64        `json:"id" db:"id"`
	OrderID       uint64       `json:"orderID" db:"order_id"`
	Exchange      ExchangeName `json:"exchange" db:"exchange"`
	Price         float64      `json:"price" db:"price"`
	Quantity      float64      `json:"quantity" db:"quantity"`
	QuoteQuantity float64      `json:"quoteQuantity" db:"quote_quantity"`
	Symbol        string       `json:"symbol" db:"symbol"`

	Side        SideType `json:"side" db:"side"`
	IsBuyer     bool     `json:"isBuyer" db:"is_buyer"`
	IsMaker     bool     `json:"isMaker" db:"is_maker"`
	Time        Time     `json:"tradedAt" db:"traded_at"`
	Fee         float64  `json:"fee" db:"fee"`
	FeeCurrency string   `json:"feeCurrency" db:"fee_currency"`

	IsMargin   bool `json:"isMargin" db:"is_margin"`
	IsIsolated bool `json:"isIsolated" db:"is_isolated"`

	StrategyID sql.NullString  `json:"strategyID" db:"strategy"`
	PnL        sql.NullFloat64 `json:"pnl" db:"pnl"`
}

func (trade Trade) String() string {
	return fmt.Sprintf("TRADE %s %s %4s %f @ %f orderID %d %s amount %f",
		trade.Exchange.String(),
		trade.Symbol,
		trade.Side,
		trade.Quantity,
		trade.Price,
		trade.OrderID,
		trade.Time.Time().Format(time.StampMilli),
		trade.QuoteQuantity)
}

// PlainText is used for telegram-styled messages
func (trade Trade) PlainText() string {
	return fmt.Sprintf("Trade %s %s %s %f @ %f, amount %f",
		trade.Exchange.String(),
		trade.Symbol,
		trade.Side,
		trade.Quantity,
		trade.Price,
		trade.QuoteQuantity)
}

func (trade Trade) Liquidity() (o string) {
	if trade.IsMaker {
		o += "MAKER"
	} else {
		o += "TAKER"
	}

	return o
}

func (trade Trade) Key() TradeKey {
	return TradeKey{ID: trade.ID, Side: trade.Side}
}

type TradeKey struct {
	ID   int64
	Side SideType
}
