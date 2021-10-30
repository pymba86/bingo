package grid

import (
	"context"
	"github.com/pymba86/bingo/pkg/engine"
	"github.com/sirupsen/logrus"
)

const Id = "grid"

var log = logrus.WithField("strategy", Id)

func init() {
	engine.RegisterStrategy(Id, &Strategy{})
}

type Strategy struct {
	Symbol string `json:"symbol"`
}

func (s *Strategy) Id() string {
	return Id
}

func (s *Strategy) Subscribe(session *engine.ExchangeSession) {
}

func (s *Strategy) Run(ctx context.Context, orderExecutor engine.OrderExecutor, session *engine.ExchangeSession) error {

	log.Infof("symbol: %s", s.Symbol)

	return nil
}
