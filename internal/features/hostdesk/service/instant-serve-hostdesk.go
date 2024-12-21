package service

import (
	"context"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/hostdesk/domain"
	"queue-bite/internal/features/hostdesk/repository"
	wld "queue-bite/internal/features/waitlist/domain"
	"queue-bite/internal/platform/eventbus"
)

var INSTANT_SERVE = "hostdesk/instant-serve"

type InstantServeHostDesk struct {
	logger   log.Logger
	repo     repository.HostDeskRepository
	eventbus eventbus.EventBus

	totalSeats int
}

func NewInstantServeHostDesk(
	logger log.Logger,
	totalSeats int,
	repo repository.HostDeskRepository,
	eventbus eventbus.EventBus,
	eventRegistry *eventbus.EventRegistry,
) *InstantServeHostDesk {

	eventRegistry.Register(domain.TopicPartyPreserved, domain.SeatsPreservedEvent{})

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

func (h *InstantServeHostDesk) NotifyPartyReady(ctx context.Context, party *wld.QueuedParty) error {
	ok, err := h.PreserveSeats(ctx, party.ID, party.Size)
	if err != nil {
		return err
	}

	if !ok {
		return domain.ErrInsufficientCapacity
	}

	h.eventbus.Publish(ctx, &domain.SeatsPreservedEvent{PartyID: party.ID})
	h.logger.LogDebug(INSTANT_SERVE, "seats preserved, notify party ready", "party id", party.ID)
	return nil
}

func (h *InstantServeHostDesk) PreserveSeats(ctx context.Context, partyID d.PartyID, seats int) (bool, error) {
	curr, err := h.repo.GetPartyServiceState(ctx, partyID)
	if err != nil {
		return false, err
	}

	// TODO: check state transition by single source of truth
	if curr != nil {
		return false, domain.ErrPartyAlreadyExists
	}

	state := domain.NewPartyServiceFromPreserve(partyID, seats)
	err = h.repo.CreatePartyServiceState(ctx, state)

	if err != nil {
		return false, err
	}
	return true, nil
}
