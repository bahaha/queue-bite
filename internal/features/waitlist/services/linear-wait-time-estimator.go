package services

import (
	"context"
	"time"

	domain "queue-bite/internal/features/waitlist"
)

type LinearWaitTimeEstimator struct {
}

func NewLinearWaitTimeEstimator() *LinearWaitTimeEstimator {
	return &LinearWaitTimeEstimator{}
}

func (e *LinearWaitTimeEstimator) EstimateWaitTime(ctx context.Context, payty *domain.Party) (time.Duration, error) {
	return 5 * time.Minute, nil
}
