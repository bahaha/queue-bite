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
	logger        log.Logger
	state         map[d.PartyID]*domain.PartyServiceState
	totalOccupied int
	mu            sync.RWMutex
}

func NewInMemoryHostDeskRepository(logger log.Logger) *InMemoryHostDeskRepository {
	return &InMemoryHostDeskRepository{
		logger:        logger,
		state:         make(map[d.PartyID]*domain.PartyServiceState),
		totalOccupied: 0,
	}
}

func (r *InMemoryHostDeskRepository) OccupySeats(ctx context.Context, partyID d.PartyID, seats int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.state[partyID]; exists {
		return domain.ErrPartyAlreadySeated
	}

	r.state[partyID] = &domain.PartyServiceState{
		ID:            partyID,
		SeatsOccupied: seats,
		Status:        domain.PartySeated,
	}
	r.totalOccupied += seats
	return nil
}

func (r *InMemoryHostDeskRepository) ReleaseSeats(ctx context.Context, partyID d.PartyID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	party, exists := r.state[partyID]
	if !exists {
		err := domain.ErrPartyNotFound
		r.logger.LogErr(INMEMORY_HOSTDESK, err, "could not release seats for party", "party id", partyID)
		return err
	}

	r.totalOccupied -= party.SeatsOccupied
	delete(r.state, partyID)
	return nil
}

func (r *InMemoryHostDeskRepository) GetOccupiedSeats(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.totalOccupied, nil
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

func (r *InMemoryHostDeskRepository) UpdatePartyServiceState(ctx context.Context, partyID d.PartyID, nextState *domain.PartyServiceState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentState, exists := r.state[partyID]
	if !exists && nextState.Status != domain.SeatReady {
		err := domain.ErrPartyNotFound
		r.logger.LogErr(INMEMORY_HOSTDESK, err, "could not update service state of party", "party id", partyID)
		return err
	}

	if currentState != nil {
		oldSeats := currentState.SeatsOccupied
		err := copier.CopyWithOption(r.state[partyID], nextState, copier.Option{
			IgnoreEmpty: true,
		})
		if err != nil {
			return err
		}

		// Update capacity if seats changed
		if nextState.SeatsOccupied != 0 && nextState.SeatsOccupied != oldSeats {
			r.totalOccupied = r.totalOccupied - oldSeats + nextState.SeatsOccupied
		}
	}

	return nil
}
