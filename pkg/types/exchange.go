package types

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const DateFormat = "2006-01-02"

type ExchangeName string

func (n ExchangeName) String() string {
	return string(n)
}

func (n *ExchangeName) Value() (driver.Value, error) {
	return n.String(), nil
}

func (n *ExchangeName) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "binance":
		*n = ExchangeName(s)
		return nil

	}

	return fmt.Errorf("unknown or unsupported exchange name: %s, valid names are: max, binance, ftx", s)
}

const (
	ExchangeBinance = ExchangeName("binance")
)

var SupportedExchanges = []ExchangeName{"binance"}

func ValidExchangeName(a string) (ExchangeName, error) {
	switch strings.ToLower(a) {
	case "binance", "bn":
		return ExchangeBinance, nil
	}

	return "", fmt.Errorf("invalid exchange name: %s", a)
}

type Exchange interface {
	Name() ExchangeName
	PlatformFeeCurrency() string

	ExchangeMarketDataService

	ExchangeTradeService
}

type ExchangeTradeService interface {
	QueryAccount(ctx context.Context) (*Account, error)

	QueryAccountBalances(ctx context.Context) (BalanceMap, error)

	SubmitOrders(ctx context.Context, orders ...SubmitOrder) (createdOrders OrderSlice, err error)

	QueryOpenOrders(ctx context.Context, symbol string) (orders []Order, err error)

	CancelOrders(ctx context.Context, orders ...Order) error
}

type ExchangeMarketDataService interface {
	NewStream() Stream

	QueryMarkets(ctx context.Context) (MarketMap, error)

	QueryTicker(ctx context.Context, symbol string) (*Ticker, error)

	QueryTickers(ctx context.Context, symbol ...string) (map[string]Ticker, error)

	QueryKLines(ctx context.Context, symbol string, interval Interval, options KLineQueryOptions) ([]KLine, error)
}

type TradeQueryOptions struct {
	StartTime   *time.Time
	EndTime     *time.Time
	Limit       int64
	LastTradeID int64
}