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

	// GetTotalSeatsInUse returns combined occupied and preserved seats with version.
	// Version enables optimistic locking for capacity changes.
	GetTotalSeatsInUse(ctx context.Context) (int, d.Version, error)

	ReleasePreservedSeats(ctx context.Context, partyID d.PartyID) error

	// TransferToOccupied moves party from preserved to occupied state.
	// Called when party checks in to start service.
	TransferToOccupied(ctx context.Context, partyID d.PartyID) error

	GetPartyServiceState(ctx context.Context, partyID d.PartyID) (*domain.PartyServiceState, error)

	// CreatePartyServiceState initializes new service state for party.
	CreatePartyServiceState(ctx context.Context, state *domain.PartyServiceState) error

	// OptimisticCreatePartyServiceState creates service state if version matches.
	// Used to handle concurrent seating operations safely.
	OptimisticCreatePartyServiceState(ctx context.Context, state *domain.PartyServiceState, version d.Version) error

	UpdatePartyServiceState(ctx context.Context, partyID d.PartyID, state *domain.PartyServiceState) error

	// EndPartyServiceState completes service and cleans up state.
	// Frees occupied seats and removes party records.
	EndPartyServiceState(ctx context.Context, partyID d.PartyID) error
}
