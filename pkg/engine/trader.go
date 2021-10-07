package engine

import (
	"context"
	"sync"
)

type SingleExchangeStrategy interface {
	Id() string
	Run(ctx context.Context) error
}

type CrossExchangeStrategy interface {
	Id() string
	CrossRun(ctx context.Context) error
}

type Trader struct {
	crossExchangeStrategies []CrossExchangeStrategy

	exchangeStrategies map[string][]SingleExchangeStrategy

	logger Logger
}

type Graceful struct {
	shutdownCallbacks []func(ctx context.Context, wg *sync.WaitGroup)
}
