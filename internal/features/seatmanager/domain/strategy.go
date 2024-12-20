package domain

import (
	"context"
	"queue-bite/internal/domain"
)

type SeatingStrategy interface {
	EvaluateNextParty(ctx context.Context, vacancySeats int) (*domain.Party, error)
}
