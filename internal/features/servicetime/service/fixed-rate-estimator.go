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

// Provides a simple, fixed-time-per-guest estimation approach suitable for uniform service scenarios.
// Accepts a `timePerGuest` parameter to define the base time per guest
//
// Example:
//
//	// This estimator will allocate 5 minutes of service time per guest.
//	estimator := NewFixedRateEstimator(5 * time.Minute)
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
