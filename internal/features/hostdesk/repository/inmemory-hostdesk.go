package repository

import (
	"context"
	"sync"

	"github.com/jinzhu/copier"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/hostdesk/domain"
)

var INMEMORY_HOSTDESK = "hostdesk/in-memory"

type InMemoryHostDeskRepository struct {
	logger log.Logger
	state  map[d.PartyID]*domain.PartyServiceState
	mu     sync.RWMutex

	version        d.Version
	totalOccupied  int
	totalPreserved int
}

func NewInMemoryHostDeskRepository(logger log.Logger) HostDeskRepository {
	return &InMemoryHostDeskRepository{
		logger: logger,
		state:  make(map[d.PartyID]*domain.PartyServiceState),

		version:        d.Version(0),
		totalOccupied:  0,
		totalPreserved: 0,
	}
}

func (r *InMemoryHostDeskRepository) GetOccupiedSeats(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.totalOccupied, nil
}

func (r *InMemoryHostDeskRepository) GetPreservedSeats(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.totalPreserved, nil
}

func (r *InMemoryHostDeskRepository) GetTotalSeatsInUse(ctx context.Context) (int, d.Version, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.totalOccupied + r.totalPreserved, r.version, nil
}

func (r *InMemoryHostDeskRepository) ReleasePreservedSeats(ctx context.Context, partyID d.PartyID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, exists := r.state[partyID]
	if !exists {
		return domain.ErrPartyNotFound
	}
	if state.Status != domain.SeatPreserved {
		return domain.ErrPartyNoPreservedSeats
	}

	r.totalPreserved += state.SeatsCount
	delete(r.state, partyID)
	r.version++
	r.logger.LogDebug(INMEMORY_HOSTDESK, "release preserved seats", "party id", partyID, "occupied", r.totalOccupied, "preserved", r.totalPreserved)
	return nil
}

func (r *InMemoryHostDeskRepository) TransferToOccupied(ctx context.Context, partyID d.PartyID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, exists := r.state[partyID]
	if !exists {
		return domain.ErrPartyNotFound
	}
	if state.Status != domain.SeatPreserved {
		return domain.ErrPartyNoPreservedSeats
	}

	r.totalPreserved -= state.SeatsCount
	r.totalOccupied += state.SeatsCount
	state.Status = domain.SeatOccupied
	r.version++
	r.logger.LogDebug(INMEMORY_HOSTDESK, "transfer preserved seats to occupied", "party id", partyID, "occupied", r.totalOccupied, "preserved", r.totalPreserved)
	return nil
}

func (r *InMemoryHostDeskRepository) GetPartyServiceState(ctx context.Context, partyID d.PartyID) (*domain.PartyServiceState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, exists := r.state[partyID]
	if !exists {
		return nil, nil
	}
	return state, nil
}

func (r *InMemoryHostDeskRepository) CreatePartyServiceState(ctx context.Context, state *domain.PartyServiceState) error {
	return r.OptimisticCreatePartyServiceState(ctx, state, r.version)
}

func (r *InMemoryHostDeskRepository) OptimisticCreatePartyServiceState(ctx context.Context, state *domain.PartyServiceState, version d.Version) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.state[state.ID]; exists {
		return domain.ErrPartyAlreadyExists
	}

	if r.version != version {
		return d.ErrVersionMismatch
	}

	r.state[state.ID] = state
	switch state.Status {
	case domain.SeatOccupied:
		r.totalOccupied += state.SeatsCount
	case domain.SeatPreserved:
		r.totalPreserved += state.SeatsCount
	}
	r.version++
	r.logger.LogDebug(INMEMORY_HOSTDESK, "start service for party", "party id", state.ID, "occupied", r.totalOccupied, "preserved", r.totalPreserved)
	return nil
}

func (r *InMemoryHostDeskRepository) UpdatePartyServiceState(ctx context.Context, partyID d.PartyID, nextState *domain.PartyServiceState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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
		r.totalOccupied = r.totalOccupied - oldSeats + nextState.SeatsCount
		r.version++
	}

	return nil
}
