package types

import (
	"database/sql"
	"fmt"
	"time"
)

type Trade struct {
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