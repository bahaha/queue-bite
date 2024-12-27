package service

import (
	"context"

	d "queue-bite/internal/domain"
	w "queue-bite/internal/features/waitlist/domain"
)

type HostDesk interface {
	GetTotalCapacity(ctx context.Context) (int, error)

	// GetCurrentCapacity returns available seats and current version.
	// Version used for optimistic locking in seat operations.
	GetCurrentCapacity(ctx context.Context) (int, d.Version, error)

	NotifyPartyReady(ctx context.Context, party *w.QueuedParty) error

	// PreserveSeats attempts to reserve seats for party.
	// Uses version for optimistic locking to handle concurrent requests.
	// Returns (true, nil) if seats successfully preserved.
	PreserveSeats(ctx context.Context, partyID d.PartyID, seats int, version d.Version) (bool, error)

	ReleasePreservedSeats(ctx context.Context, partyID d.PartyID) (bool, error)

	// ServeImmediately seats party without going through queue.
	// Used when capacity is immediately available.
	ServeImmediately(ctx context.Context, party *d.Party) error

	// CheckIn confirms party arrival and starts service.
	// Transitions from preserved seats to occupied status.
	CheckIn(ctx context.Context, party *w.QueuedParty) error

	// ServiceComplete ends party's service and frees seats.
	// Triggers capacity updates and cleanup.
	ServiceComplete(ctx context.Context, party *w.QueuedParty) error

	HasPartyOccupiedSeat(ctx context.Context, partyID d.PartyID) bool
}
