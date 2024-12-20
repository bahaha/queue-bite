package service

import (
	"context"
	"time"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/hostdesk/domain"
	"queue-bite/internal/features/hostdesk/repository"
	"queue-bite/internal/platform/eventbus"
)

var INSTANT_SERVE = "hostdesk/instant-serve"

type InstantServeHostDesk struct {
	logger     log.Logger
	totalSeats int
	repo       repository.HostDeskRepository
	eventbus   eventbus.EventBus
}

func NewInstantServeHostDesk(
	logger log.Logger,
	totalSeats int,
	repo repository.HostDeskRepository,
	eventbus eventbus.EventBus,
	eventRegistry *eventbus.EventRegistry,
) *InstantServeHostDesk {

	eventRegistry.Register(domain.TopicPartyReady, domain.PartyReadyEvent{})

	return &InstantServeHostDesk{
		logger:     logger,
		totalSeats: totalSeats,
		repo:       repo,
		eventbus:   eventbus,
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
	if err := h.repo.UpdatePartyServiceState(ctx, partyID, &domain.PartyServiceState{
		Status:     domain.SeatReady,
		NotifiedAt: time.Now(),
	}); err != nil {
		h.logger.LogErr(INSTANT_SERVE, err, "update party state when party are ready to serve", "party id", partyID)
		return err
	}

	h.eventbus.Publish(ctx, &domain.PartyReadyEvent{
		PartyID: partyID,
		ReadyAt: time.Now(),
	})
	h.logger.LogDebug(INSTANT_SERVE, "notify party ready event", "party id", partyID, "ready at", time.Now())
	return nil
}

func (h *InstantServeHostDesk) GetPartyServiceState(ctx context.Context, partyID d.PartyID) (*domain.PartyServiceState, error) {
	return h.repo.GetPartyServiceState(ctx, partyID)
}
