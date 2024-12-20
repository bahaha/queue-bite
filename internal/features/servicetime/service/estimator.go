package service

import (
	"context"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/servicetime/domain"
)

type ServiceTimeEstimator interface {
	EstimateServiceTime(ctx context.Context, party *d.Party) (*domain.ServiceTimeEstimate, error)
}
