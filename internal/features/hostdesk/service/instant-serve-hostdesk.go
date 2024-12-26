package service

import (
	"context"
	"fmt"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/hostdesk/domain"
	"queue-bite/internal/features/hostdesk/repository"
	wld "queue-bite/internal/features/waitlist/domain"
	"queue-bite/internal/platform/eventbus"
)

var INSTANT_SERVE = "hostdesk/instant-serve"
var SKIP_VERSION_CHECK = d.Version(-1)

type InstantServeHostDesk struct {
	logger       log.Logger
	repo         repository.HostDeskRepository
	eventbus     eventbus.EventBus
	servicetimer ServiceTimer
	totalSeats   int
}

func NewInstantServeHostDesk(
	logger log.Logger,
	totalSeats int,
	repo repository.HostDeskRepository,
	eventbus eventbus.EventBus,
	servicetimer ServiceTimer,
) HostDesk {
	return &InstantServeHostDesk{
		logger:       logger,
		totalSeats:   totalSeats,
		repo:         repo,
		eventbus:     eventbus,
		servicetimer: servicetimer,
	}
}

func (h *InstantServeHostDesk) GetTotalCapacity(ctx context.Context) (int, error) {
	return h.totalSeats, nil
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
	preserved, err := h.PreserveSeats(ctx, party.ID, party.Size, SKIP_VERSION_CHECK)
	if err != nil {
		return err
	}

	if preserved {
		h.eventbus.Publish(ctx, &domain.SeatsPreservedEvent{PartyID: party.ID})
		h.logger.LogDebug(INSTANT_SERVE, "seats preserved, notify party ready", "party id", party.ID)
	}
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

	cap, v, err := h.GetCurrentCapacity(ctx)
	if version != SKIP_VERSION_CHECK && v != version {
		return false, d.ErrVersionMismatch
	}

	if cap < seats {
		return false, domain.ErrInsufficientCapacity
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
	if party.Status == d.PartyStatusReady {
		if err := h.repo.TransferToOccupied(ctx, party.ID); err != nil {
			return err
		}
		h.logger.LogDebug(INSTANT_SERVE, "party checked in, transfer preserved seats to occupied", "party", party)
	} else if party.Status == d.PartyStatusServing {
		h.logger.LogDebug(INSTANT_SERVE, "party checkin in, start serving", "party", party)
	} else {
		return fmt.Errorf("invalid state of party checkin: %v", party)
	}

	if h.servicetimer != nil {
		h.servicetimer.StartTracking(ctx, party, func(ctx context.Context, partyID d.PartyID) error {
			return h.ServiceComplete(ctx, party)
		})
	}
	return nil
}

func (h *InstantServeHostDesk) ServiceComplete(ctx context.Context, party *wld.QueuedParty) error {
	if err := h.repo.EndPartyServiceState(ctx, party.ID); err != nil {
		return err
	}
	h.logger.LogDebug(INSTANT_SERVE, "service completed", "party", party)
	if err := h.eventbus.Publish(ctx, domain.PartyServiceCompeletedEvent{PartyID: party.ID}); err != nil {
		h.logger.LogErr(INSTANT_SERVE, err, "could not publish service completed event")
		return err
	}
	h.logger.LogDebug(INSTANT_SERVE, "publish party service completed event", "party id", party.ID)
	return nil
}

func (h *InstantServeHostDesk) HasPartyOccupiedSeat(ctx context.Context, partyID d.PartyID) bool {
	state, err := h.repo.GetPartyServiceState(ctx, partyID)
	h.logger.LogDebug(INSTANT_SERVE, "get party service state", "party", state)
	if err != nil || state == nil {
		return false
	}

	return state.Status == domain.SeatOccupied
}
