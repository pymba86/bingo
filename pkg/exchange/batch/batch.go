package batch

import (
	"context"
	"github.com/pymba86/bingo/pkg/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"time"
)

type TradeBatchQuery struct {
	types.Exchange
}

func (e TradeBatchQuery) Query(ctx context.Context, symbol string, options *types.TradeQueryOptions) (c chan types.Trade, errC chan error) {
	c = make(chan types.Trade, 500)
	errC = make(chan error, 1)

	tradeHistoryService, ok := e.Exchange.(types.ExchangeTradeHistoryService)
	if !ok {
		// skip exchanges that does not support trading history services
		logrus.Warnf(
			"exchange %s does not implement ExchangeTradeHistoryService, skip syncing closed orders",
			e.Exchange.Name())
		return c, errC
	}

	var lastTradeID = options.LastTradeID

	go func() {
		limiter := rate.NewLimiter(rate.Every(5*time.Second), 2) // from binance (original 1200, use 1000 for safety)

		defer close(c)
		defer close(errC)

		var tradeKeys = map[types.TradeKey]struct{}{}

		for {
			if err := limiter.Wait(ctx); err != nil {
				logrus.WithError(err).Error("rate limit error")
			}

			logrus.Infof("querying %s trades from id=%d limit=%d", symbol, lastTradeID, options.Limit)

			var err error
			var trades []types.Trade

			trades, err = tradeHistoryService.QueryTrades(ctx, symbol, &types.TradeQueryOptions{
				Limit:       options.Limit,
				LastTradeID: lastTradeID,
			})

			if err != nil {
				errC <- err
				return
			}

			if len(trades) == 0 {
				return
			} else if len(trades) == 1 {
				k := trades[0].Key()
				if _, exists := tradeKeys[k]; exists {
					return
				}
			}

			for _, t := range trades {
				key := t.Key()
				if _, ok := tradeKeys[key]; ok {
					logrus.Debugf("ignore duplicated trade: %+v", key)
					continue
				}

				lastTradeID = t.ID
				tradeKeys[key] = struct{}{}

				// ignore the first trade if last TradeID is given
				c <- t
			}
		}
	}()

	return c, errC
}
