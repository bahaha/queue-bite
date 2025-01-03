package repository

import (
	"context"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/waitlist/domain"
)

// WaitlistRepository defines the persistence operations for waitlist management.
// It handles the storage and retrieval of queued parties and their associated data.
type WaitlistRepositoy interface {
	// HasParty return if a party is in waitlist queue
	HasParty(ctx context.Context, partyID d.PartyID) bool

	// AddParty stores a new party in the queue and calculates their position
	// and waiting time based on parties ahead of them.
	AddParty(ctx context.Context, party *domain.QueuedParty) (*domain.QueuedParty, error)

	// RemoveParty removes a party from the queue and updates wait times
	// for parties behind them in the queue.
	RemoveParty(ctx context.Context, partyID d.PartyID) error

	// GetParty retrieves a party's current queue information.
	// Returns nil, nil if party is not found.
	GetParty(ctx context.Context, partyID d.PartyID) (*domain.QueuedParty, error)

	// GetPartyDetails retrieves a party's detail without the queue information,
	// this operation should be a quick lookup without checking queue order like the GetParty does
	// Returns nil, nil if party is not found.
	GetPartyDetails(ctx context.Context, partyID d.PartyID) (*domain.QueuedParty, error)

	// GetQueueStatus retrieves current queue metrics.
	GetQueueStatus(ctx context.Context) (*domain.QueueStatus, error)

	// ScanParties streams queued parties in order through a channel.
	// Uses batched retrieval (ZRANGE) with configured scanRange to manage memory usage.
	// Streaming provides efficient iteration over large queues without loading all at once.
	//
	// Usage:
	//   parties, err := repo.ScanParties(ctx)
	//   if err != nil {
	//       return err
	//   }
	//   for party := range parties {
	//       // Process each party
	//   }
	//
	// The channel is closed when:
	//  - All parties have been streamed
	//  - Context is cancelled
	//  - Error occurs during scanning
	ScanParties(ctx context.Context) (<-chan *domain.QueuedParty, error)

	// UpdatePartyStatus update a party's current state in the queue.
	// Returns nil, nil if party is not found.
	UpdatePartyStatus(ctx context.Context, partyID d.PartyID, status d.PartyStatus) error
}
