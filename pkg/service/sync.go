package service

import (
	"context"
	"github.com/pymba86/bingo/pkg/types"
	"time"
)

type SyncService struct {
	TradeService *TradeService
}

func (s *SyncService) SyncSessionSymbols(ctx context.Context, exchange types.Exchange,
	startTime time.Time, symbols ...string) error {

	for _, symbol := range symbols {
		if err := s.TradeService.Sync(ctx, exchange, symbol); err != nil {
			return err
		}
	}

	return nil
}
