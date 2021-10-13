package engine

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
)

type SingleExchangeStrategy interface {
	Id() string
	Run(ctx context.Context, orderExecutor OrderExecutor, session *ExchangeSession) error
}

type CrossExchangeStrategy interface {
	Id() string
	CrossRun(ctx context.Context, orderExecutionRouter OrderExecutionRouter, sessions map[string]*ExchangeSession) error
}

type ExchangeSessionSubscriber interface {
	Subscribe(session *ExchangeSession)
}

type CrossExchangeSessionSubscriber interface {
	CrossSubscribe(sessions map[string]*ExchangeSession)
}

type Trader struct {
	environment *Environment

	crossExchangeStrategies []CrossExchangeStrategy

	exchangeStrategies map[string][]SingleExchangeStrategy

	logger Logger

	graceful Graceful
}

func (trader *Trader) Configure(config *Config) error {
	return nil
}

func (trader *Trader) Run(ctx context.Context) error {

	trader.Subscribe()



	return nil
}

func (trader *Trader) Subscribe() {
	for sessionName, strategies := range trader.exchangeStrategies {
		session := trader.environment.sessions[sessionName]
		for _, strategy := range strategies {
			if subscriber, ok := strategy.(ExchangeSessionSubscriber); ok {
				subscriber.Subscribe(session)
			} else {
				log.Errorf("strategy %s does not implement ExchangeSessionSubscriber", strategy.Id())
			}
		}
	}

	for _, strategy := range trader.crossExchangeStrategies {
		if subscriber, ok := strategy.(CrossExchangeSessionSubscriber); ok {
			subscriber.CrossSubscribe(trader.environment.sessions)
		} else {
			log.Errorf("strategy %s does not implement CrossExchangeSessionSubscriber", strategy.Id())
		}
	}
}

func (trader *Trader) AttachStrategyOn(session string, strategies ...SingleExchangeStrategy) error {

	if len(trader.environment.sessions) == 0 {
		return fmt.Errorf(
			"you don't have any session configured, please check your environment variable or config file")
	}

	if _, ok := trader.environment.sessions[session]; !ok {
		var keys []string
		for k := range trader.environment.sessions {
			keys = append(keys, k)
		}

		return fmt.Errorf("session %s is not defined, valid sessions are: %v", session, keys)
	}

	for _, s := range strategies {
		trader.exchangeStrategies[session] = append(trader.exchangeStrategies[session], s)
	}

	return nil
}

func (trader *Trader) AttachCrossExchangeStrategy(strategy CrossExchangeStrategy) *Trader {

	trader.crossExchangeStrategies = append(
		trader.crossExchangeStrategies, strategy)

	return trader
}

type Graceful struct {
	shutdownCallbacks []func(ctx context.Context, wg *sync.WaitGroup)
}
