package repository

import (
	"context"

	d "queue-bite/internal/domain"
	"queue-bite/internal/features/hostdesk/domain"
)

// HostDeskRepository defines the interface for managing seating state.
// It tracks occupied seats and party service states to help the host manage
// restaurant capacity efficiently.
type HostDeskRepository interface {
	// OccupySeats marks the specified number of seats as occupied for a party.
	// Returns ErrInsufficientCapacity if there aren't enough available seats.
	OccupySeats(ctx context.Context, partyID d.PartyID, seats int) error

	// ReleaseSeats frees up the seats that were occupied by the specified party.
	// Returns ErrPartyNotFound if the party isn't currently occupying any seats.
	ReleaseSeats(ctx context.Context, partyID d.PartyID) error

	// GetOccupiedSeats returns the total number of currently occupied seats.
	GetOccupiedSeats(ctx context.Context) (int, error)

	// GetPartyServiceState retrieves the current service state for a party.
	// Returns nil if the party isn't found.
	GetPartyServiceState(ctx context.Context, partyID d.PartyID) (*domain.PartyServiceState, error)

	// UpdatePartyServiceState updates the service state for a party.
	// Returns ErrPartyNotFound if the party isn't currently being served.
	UpdatePartyServiceState(ctx context.Context, partyID d.PartyID, state *domain.PartyServiceState) error
}
