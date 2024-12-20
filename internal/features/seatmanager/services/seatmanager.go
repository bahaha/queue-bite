package services

import (
	"context"
	"queue-bite/internal/features/seatmanager/domain"
)

type seatManager struct {
	strategy domain.SeatingStrategy
	stopCh   chan struct{}
}

func NewSeatManager(
	strategy domain.SeatingStrategy,
	stopCh chan struct{},
) *seatManager {
	return &seatManager{
		strategy: strategy,
		stopCh:   stopCh,
	}
}

func (*seatManager) WatchSeatVacancy(ctx context.Context) error {
	return nil
}

func (*seatManager) UnwatchSeatVacancy(ctx context.Context) error {
	return nil
}

func (*seatManager) HandleCapacityAvailable(ctx context.Context, event interface{}) error {
	return nil
}

func (*seatManager) HandleServiceCompleted(ctx context.Context, event interface{}) error {
	return nil
}
