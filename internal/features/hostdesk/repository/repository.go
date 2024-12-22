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
	GetOccupiedSeats(ctx context.Context) (int, error)

	GetPreservedSeats(ctx context.Context) (int, error)

	GetTotalSeatsInUse(ctx context.Context) (int, d.Version, error)

	ReleasePreservedSeats(ctx context.Context, partyID d.PartyID) error

	TransferToOccupied(ctx context.Context, partyID d.PartyID) error

	GetPartyServiceState(ctx context.Context, partyID d.PartyID) (*domain.PartyServiceState, error)

	CreatePartyServiceState(ctx context.Context, state *domain.PartyServiceState) error

	OptimisticCreatePartyServiceState(ctx context.Context, state *domain.PartyServiceState, version d.Version) error

	UpdatePartyServiceState(ctx context.Context, partyID d.PartyID, state *domain.PartyServiceState) error
}
