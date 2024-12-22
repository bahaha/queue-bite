package repository

import (
	"context"
	"sync/atomic"

	"github.com/jinzhu/copier"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/hostdesk/domain"
)

var INMEMORY_HOSTDESK = "hostdesk/in-memory"

type hostdeskStats struct {
	Occupied  int
	Preserved int
	Version   d.Version
}

type InMemoryHostDeskRepository struct {
	logger log.Logger
	state  map[d.PartyID]*domain.PartyServiceState
	stats  atomic.Value
}

func NewInMemoryHostDeskRepository(logger log.Logger) HostDeskRepository {
	repo := &InMemoryHostDeskRepository{
		logger: logger,
		state:  make(map[d.PartyID]*domain.PartyServiceState),
	}
	repo.stats.Store(hostdeskStats{
		Occupied:  0,
		Preserved: 0,
		Version:   0,
	})
	return repo
}

func (r *InMemoryHostDeskRepository) GetOccupiedSeats(ctx context.Context) (int, error) {
	state := r.stats.Load().(hostdeskStats)
	return state.Occupied, nil
}

func (r *InMemoryHostDeskRepository) GetPreservedSeats(ctx context.Context) (int, error) {
	state := r.stats.Load().(hostdeskStats)
	return state.Preserved, nil
}

func (r *InMemoryHostDeskRepository) GetTotalSeatsInUse(ctx context.Context) (int, d.Version, error) {
	state := r.stats.Load().(hostdeskStats)
	return state.Occupied + state.Preserved, d.Version(state.Version), nil
}

func (r *InMemoryHostDeskRepository) ReleasePreservedSeats(ctx context.Context, partyID d.PartyID) error {
	state, exists := r.state[partyID]
	if !exists {
		return domain.ErrPartyNotFound
	}
	if state.Status != domain.SeatPreserved {
		return domain.ErrPartyNoPreservedSeats
	}

	stats := r.stats.Load().(hostdeskStats)
	newStats := hostdeskStats{
		Occupied:  stats.Occupied,
		Preserved: stats.Preserved + state.SeatsCount,
		Version:   stats.Version + 1,
	}
	r.stats.Store(newStats)

	delete(r.state, partyID)
	r.logger.LogDebug(INMEMORY_HOSTDESK, "release preserved seats", "party id", partyID, "stats", newStats)
	return nil
}

func (r *InMemoryHostDeskRepository) TransferToOccupied(ctx context.Context, partyID d.PartyID) error {
	state, exists := r.state[partyID]
	if !exists {
		return domain.ErrPartyNotFound
	}
	if state.Status != domain.SeatPreserved {
		return domain.ErrPartyNoPreservedSeats
	}

	stats := r.stats.Load().(hostdeskStats)
	nextStats := hostdeskStats{
		Occupied:  stats.Occupied + state.SeatsCount,
		Preserved: stats.Preserved - state.SeatsCount,
		Version:   stats.Version + 1,
	}
	r.stats.Store(nextStats)
	state.Status = domain.SeatOccupied

	r.logger.LogDebug(INMEMORY_HOSTDESK, "transfer preserved seats to occupied", "party id", partyID, "stats", nextStats)
	return nil
}

func (r *InMemoryHostDeskRepository) GetPartyServiceState(ctx context.Context, partyID d.PartyID) (*domain.PartyServiceState, error) {
	state, exists := r.state[partyID]
	if !exists {
		return nil, nil
	}
	return state, nil
}

func (r *InMemoryHostDeskRepository) CreatePartyServiceState(ctx context.Context, state *domain.PartyServiceState) error {
	stats := r.stats.Load().(hostdeskStats)
	return r.OptimisticCreatePartyServiceState(ctx, state, stats.Version)
}

func (r *InMemoryHostDeskRepository) OptimisticCreatePartyServiceState(ctx context.Context, state *domain.PartyServiceState, version d.Version) error {
	if _, exists := r.state[state.ID]; exists {
		return domain.ErrPartyAlreadyExists
	}

	stats := r.stats.Load().(hostdeskStats)
	if stats.Version != version {
		return d.ErrVersionMismatch
	}

	r.state[state.ID] = state
	nextStats := hostdeskStats{
		Occupied:  stats.Occupied,
		Preserved: stats.Preserved,
		Version:   stats.Version + 1,
	}
	switch state.Status {
	case domain.SeatOccupied:
		nextStats.Occupied += state.SeatsCount
	case domain.SeatPreserved:
		nextStats.Preserved += state.SeatsCount
	}
	r.stats.Store(nextStats)

	r.logger.LogDebug(INMEMORY_HOSTDESK, "start service for party", "party id", state.ID, "stats", nextStats)
	return nil
}

func (r *InMemoryHostDeskRepository) UpdatePartyServiceState(ctx context.Context, partyID d.PartyID, nextState *domain.PartyServiceState) error {
	currentState, exists := r.state[partyID]
	if !exists {
		return domain.ErrPartyNotFound
	}

	oldSeats := currentState.SeatsCount
	err := copier.CopyWithOption(r.state[partyID], nextState, copier.Option{
		IgnoreEmpty: true,
	})
	if err != nil {
		return err
	}

	if nextState.SeatsCount != 0 && nextState.SeatsCount != oldSeats {
		stats := r.stats.Load().(hostdeskStats)
		r.stats.Store(hostdeskStats{
			Occupied:  stats.Occupied - oldSeats + nextState.SeatsCount,
			Preserved: stats.Preserved,
			Version:   stats.Version + 1,
		})
	}

	return nil
}
