package engine

import (
	"context"
	"fmt"
	"time"
)

var LoadedExchangeStrategies = make(map[string]SingleExchangeStrategy)
var LoadedCrossExchangeStrategies = make(map[string]CrossExchangeStrategy)

func RegisterStrategy(key string, s interface{}) {
	loaded := 0
	if d, ok := s.(SingleExchangeStrategy); ok {
		LoadedExchangeStrategies[key] = d
		loaded++
	}

	if d, ok := s.(CrossExchangeStrategy); ok {
		LoadedCrossExchangeStrategies[key] = d
		loaded++
	}

	if loaded == 0 {
		panic(fmt.Errorf("%T does not implement SingleExchangeStrategy or CrossExchangeStrategy", s))
	}
}

type Environment struct {
	Notifiability

	startTime time.Time

	sessions map[string]*ExchangeSession
}

func (e *Environment) Start(ctx context.Context) error {
	return nil
}

func (e *Environment) Connect(ctx context.Context) error {
	return nil
}
