package service

import (
	"context"
	"time"

	d "queue-bite/internal/domain"
	"queue-bite/internal/features/hostdesk/domain"
	"queue-bite/internal/features/hostdesk/repository"
)

type InstantServeHostDesk struct {
	totalSeats int
	repo       repository.HostDeskRepository
}

func NewInstantServeHostDesk(totalSeats int, repo repository.HostDeskRepository) *InstantServeHostDesk {
	return &InstantServeHostDesk{
		totalSeats: totalSeats,
		repo:       repo,
	}
}

func (h *InstantServeHostDesk) GetCurrentCapacity(ctx context.Context) (int, error) {
	occupied, err := h.repo.GetOccupiedSeats(ctx)
	if err != nil {
		return 0, err
	}

	return h.totalSeats - occupied, nil
}

func (h *InstantServeHostDesk) NotifyPartyReady(ctx context.Context, partyID d.PartyID) error {
	return h.repo.UpdatePartyServiceState(ctx, partyID, &domain.PartyServiceState{
		Status:     domain.SeatReady,
		NotifiedAt: time.Now(),
	})
}
func (h *InstantServeHostDesk) GetPartyServiceState(ctx context.Context, partyID d.PartyID) (*domain.PartyServiceState, error) {
	return h.repo.GetPartyServiceState(ctx, partyID)
}
