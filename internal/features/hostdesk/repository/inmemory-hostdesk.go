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

	totalOccupied  int
	totalPreserved int
}

func NewInMemoryHostDeskRepository(logger log.Logger) *InMemoryHostDeskRepository {
	return &InMemoryHostDeskRepository{
		logger: logger,
		state:  make(map[d.PartyID]*domain.PartyServiceState),

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
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.state[state.ID]; exists {
		return domain.ErrPartyAlreadyExists
	}

	r.state[state.ID] = state
	return nil
}

func (r *InMemoryHostDeskRepository) UpdatePartyServiceState(ctx context.Context, partyID d.PartyID, nextState *domain.PartyServiceState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentState, exists := r.state[partyID]
	if !exists {
		r.state[partyID] = nextState
		return nil
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
	}

	return nil
}
