package repository

import (
	"context"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/waitlist/domain"
)

// WaitlistRepository defines the persistence operations for waitlist management.
// It handles the storage and retrieval of queued parties and their associated data.
type WaitlistRepositoy interface {
	// AddParty stores a new party in the queue and calculates their position
	// and waiting time based on parties ahead of them.
	AddParty(ctx context.Context, party *domain.QueuedParty) (*domain.QueuedParty, error)

	// RemoveParty removes a party from the queue and updates wait times
	// for parties behind them in the queue.
	RemoveParty(ctx context.Context, partyID d.PartyID) error

	// GetParty retrieves a party's current queue information.
	// Returns nil, nil if party is not found.
	GetParty(ctx context.Context, partyID d.PartyID) (*domain.QueuedParty, error)

	// GetQueueStatus retrieves current queue metrics.
	GetQueueStatus(ctx context.Context) (*domain.QueueStatus, error)
}
