package domain

import (
	"context"
	waitlist "queue-bite/internal/features/waitlist/domain"
)

type SeatingStrategy interface {
	EvaluateNextParty(ctx context.Context, vacancySeats int) (*waitlist.QueuedParty, error)
}
