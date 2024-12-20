package services

import (
	"context"
	"queue-bite/internal/domain"
	"time"
)

type TimerBasedStrategy struct {
	serviceDuration time.Duration
}

func (s *TimerBasedStrategy) EvaluateNextParty(ctx context.Context, vacancySeats int) (*domain.Party, error) {
	return nil, nil
}
