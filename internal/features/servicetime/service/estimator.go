package service

import (
	"context"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/servicetime/domain"
)

// ServiceTimeEstimator calculates expected service duration for parties.
// Different implementations can use various factors to estimate time:
type ServiceTimeEstimator interface {

	// EstimateServiceTime returns expected service duration for a party.
	// Used for queue wait time calculations and capacity planning.
	EstimateServiceTime(ctx context.Context, party *d.Party) (*domain.ServiceTimeEstimate, error)
}
