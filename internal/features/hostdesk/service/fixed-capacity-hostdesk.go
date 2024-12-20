package service

import (
	"context"
	"queue-bite/internal/features/hostdesk/domain"
	"queue-bite/internal/features/hostdesk/repository"
)

type fixedCapacityHostDesk struct {
	totalSeats int
	repo       repository.HostDeskRepository
}

func NewFixedCapacityHostDesk(totalSeats int, repo repository.HostDeskRepository) *fixedCapacityHostDesk {
	return &fixedCapacityHostDesk{
		totalSeats: totalSeats,
		repo:       repo,
	}
}

func (h *fixedCapacityHostDesk) GetCurrentCapacity(ctx context.Context) (int, error) {
	return 0, nil
}
func (h *fixedCapacityHostDesk) NotifyPartyReady(ctx context.Context, partyID string) error {
	return nil
}
func (h *fixedCapacityHostDesk) GetPartyServiceState(ctx context.Context, partyID string) (*domain.PartyServiceState, error) {
	return nil, nil
}
