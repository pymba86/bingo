package binance

import "github.com/pymba86/bingo/pkg/types"

type Exchange struct {
	key    string
	secret string
}

func (e Exchange) Name() types.ExchangeName {
	return types.ExchangeBinance
}
