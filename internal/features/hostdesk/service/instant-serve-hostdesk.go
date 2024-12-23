package service

import (
	"context"
	"fmt"
	"time"

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
) HostDesk {
	return &InstantServeHostDesk{
		logger:     logger,
		totalSeats: totalSeats,
		repo:       repo,
		eventbus:   eventbus,
	}
}

func (h *InstantServeHostDesk) GetCurrentCapacity(ctx context.Context) (int, d.Version, error) {
	totalUsed, version, err := h.repo.GetTotalSeatsInUse(ctx)
	if err != nil {
		return h.totalSeats, version, err
	}

	capacity := h.totalSeats - totalUsed
	h.logger.LogDebug(INSTANT_SERVE, "current capacity", "capacity", capacity, "total used", totalUsed)
	return capacity, version, nil
}

func (h *InstantServeHostDesk) NotifyPartyReady(ctx context.Context, party *wld.QueuedParty) error {
	h.eventbus.Publish(ctx, &domain.SeatsPreservedEvent{PartyID: party.ID})
	h.logger.LogDebug(INSTANT_SERVE, "seats preserved, notify party ready", "party id", party.ID)
	return nil
}

func (h *InstantServeHostDesk) PreserveSeats(ctx context.Context, partyID d.PartyID, seats int, version d.Version) (bool, error) {
	curr, err := h.repo.GetPartyServiceState(ctx, partyID)
	if err != nil {
		return false, err
	}

	// TODO: check state transition by single source of truth
	if curr != nil {
		return false, domain.ErrPartyAlreadyExists
	}

	state := domain.NewPartyServiceFromPreserve(partyID, seats)
	err = h.repo.OptimisticCreatePartyServiceState(ctx, state, version)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (h *InstantServeHostDesk) ReleasePreservedSeats(ctx context.Context, partyID d.PartyID) (bool, error) {
	err := h.repo.ReleasePreservedSeats(ctx, partyID)
	if err == nil {
		return true, nil
	}
	switch err {
	case domain.ErrPartyNotFound:
	case domain.ErrPartyNoPreservedSeats:
		return false, nil
	}
	return false, fmt.Errorf("failed to release preserved seats for party: %v", partyID)
}

func (h *InstantServeHostDesk) ServeImmediately(ctx context.Context, party *d.Party) error {
	state := domain.NewPartyServiceImmediately(party.ID, party.Size)
	return h.repo.CreatePartyServiceState(ctx, state)
}

func (h *InstantServeHostDesk) CheckIn(ctx context.Context, party *wld.QueuedParty) error {
	if err := h.repo.TransferToOccupied(ctx, party.ID); err != nil {
		return err
	}

	return h.repo.UpdatePartyServiceState(ctx, party.ID, &domain.PartyServiceState{
		CheckedInAt: time.Now(),
	})
}
