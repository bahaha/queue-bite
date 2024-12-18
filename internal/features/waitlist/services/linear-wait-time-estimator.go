package services

import (
	"context"
	"time"

	domain "queue-bite/internal/features/waitlist"
)

type LinearServiceTimeEstimator struct {
}

func NewLinearWaitTimeEstimator() *LinearServiceTimeEstimator {
	return &LinearServiceTimeEstimator{}
}

func (e *LinearServiceTimeEstimator) EstimateServiceTime(ctx context.Context, payty *domain.Party) (time.Duration, error) {
	return 5 * time.Minute, nil
}
