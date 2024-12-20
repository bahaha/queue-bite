package service

import (
	"context"
	"time"

	d "queue-bite/internal/domain"
	"queue-bite/internal/features/servicetime/domain"
)

type FixedRateEstimator struct {
	timePerGuest time.Duration
}

func NewFixedRateEstimator(timePerGuest time.Duration) *FixedRateEstimator {
	return &FixedRateEstimator{
		timePerGuest: timePerGuest,
	}
}

func (e *FixedRateEstimator) EstimateServiceTime(ctx context.Context, party *d.Party) (*domain.ServiceTimeEstimate, error) {
	return &domain.ServiceTimeEstimate{
		Duration: e.timePerGuest * time.Duration(party.Size),
	}, nil
}
