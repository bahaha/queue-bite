package repository

import (
	"context"

	d "queue-bite/internal/domain"
	"queue-bite/internal/features/hostdesk/domain"

	"github.com/jinzhu/copier"
)

type InMemoryHostDeskRepository struct {
	state map[d.PartyID]*domain.PartyServiceState
}

func NewInMemoryHostDeskRepository() *InMemoryHostDeskRepository {
	return &InMemoryHostDeskRepository{
		state: make(map[d.PartyID]*domain.PartyServiceState),
	}
}

func (r *InMemoryHostDeskRepository) OccupySeats(ctx context.Context, partyID d.PartyID, seats int) error {
	if _, exists := r.state[partyID]; exists {
		return domain.ErrPartyAlreadySeated
	}

	r.state[partyID] = &domain.PartyServiceState{
		ID:            partyID,
		SeatsOccupied: seats,
		Status:        domain.PartySeated,
	}
	return nil
}

func (r *InMemoryHostDeskRepository) ReleaseSeats(ctx context.Context, partyID d.PartyID) error {
	if _, exists := r.state[partyID]; !exists {
		return domain.ErrPartyAlreadySeated
	}

	delete(r.state, partyID)
	return nil
}

func (r *InMemoryHostDeskRepository) GetOccupiedSeats(ctx context.Context) (int, error) {
	total := 0
	for _, party := range r.state {
		total += party.SeatsOccupied
	}
	return total, nil
}

func (r *InMemoryHostDeskRepository) GetPartyServiceState(ctx context.Context, partyID d.PartyID) (*domain.PartyServiceState, error) {
	state, exists := r.state[partyID]
	if !exists {
		return nil, nil
	}
	return state, nil
}

func (r *InMemoryHostDeskRepository) UpdatePartyServiceState(ctx context.Context, partyID d.PartyID, nextState *domain.PartyServiceState) error {
	if _, exists := r.state[partyID]; !exists {
		return domain.ErrPartyNotFound
	}

	copier.CopyWithOption(r.state[partyID], nextState, copier.Option{
		IgnoreEmpty: true,
	})
	return nil
}
