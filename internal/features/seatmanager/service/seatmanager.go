package service

import (
	"context"
	"fmt"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	hdd "queue-bite/internal/features/hostdesk/domain"
	hostdesk "queue-bite/internal/features/hostdesk/service"
	"queue-bite/internal/features/seatmanager/domain"
	w "queue-bite/internal/features/waitlist/domain"
	waitlist "queue-bite/internal/features/waitlist/service"
	"queue-bite/internal/platform/eventbus"

	"github.com/jinzhu/copier"
)

var SEAT_MANAGER = "seatmanager"

type SeatManager interface {
	WatchSeatVacancy(ctx context.Context) error
	UnwatchSeatVacancy(ctx context.Context) error

	ProcessNewParty(ctx context.Context, party *d.Party) (*w.QueuedParty, error)
}

type PartySelectionStrategy interface {
	EvaluateNextParty(ctx context.Context, vacancySeats int) (*w.QueuedParty, error)
}

type seatManager struct {
	logger     log.Logger
	eventbus   eventbus.EventBus
	waitlist   waitlist.Waitlist
	hostdesk   hostdesk.HostDesk
	processing PartyProcessingStrategy
	selection  PartySelectionStrategy
}

func NewSeatManager(
	logger log.Logger,
	eventbus eventbus.EventBus,
	waitlist waitlist.Waitlist,
	hostdesk hostdesk.HostDesk,
	processing PartyProcessingStrategy,
	selection PartySelectionStrategy,
) SeatManager {
	return &seatManager{
		logger:     logger,
		eventbus:   eventbus,
		waitlist:   waitlist,
		hostdesk:   hostdesk,
		processing: processing,
		selection:  selection,
	}
}

func (m *seatManager) WatchSeatVacancy(ctx context.Context) error {
	m.eventbus.Subscribe(hdd.TopicPartyPreserved, m.handleSeatPreservedEvent)
	return nil
}

func (m *seatManager) UnwatchSeatVacancy(ctx context.Context) error {
	m.logger.LogDebug(SEAT_MANAGER, "Seat manager stop observing")
	return nil
}

func (m *seatManager) ProcessNewParty(ctx context.Context, party *d.Party) (*w.QueuedParty, error) {
	capacity, err := m.hostdesk.GetCurrentCapacity(ctx)
	if err != nil {
		m.logger.LogErr(SEAT_MANAGER, err, "failed to get current capacity")
		return nil, err
	}

	queueStatus, err := m.waitlist.GetQueueStatus(ctx)
	if err != nil {
		m.logger.LogErr(SEAT_MANAGER, err, "failed to get current waitlist status")
		return nil, err
	}

	seatingCtx := &SeatingContext{
		SeatsAvailable: capacity >= party.Size,
		QueueStatus:    queueStatus,
	}
	newPartyStatus, shouldPreserve := m.processing.DeterminePartyState(ctx, seatingCtx)
	m.logger.LogDebug(SEAT_MANAGER, "determine new party should wait or serve", "seating ctx", seatingCtx, "new party stats", newPartyStatus, "should preserve", shouldPreserve)

	var needReleaseSeats bool
	defer func() {
		if needReleaseSeats {
			_, err := m.hostdesk.ReleasePreservedSeats(ctx, party.ID)
			if err != nil {
				m.logger.LogErr(SEAT_MANAGER, err, "failed release preserved seats on processing new party")
			}
		}
	}()

	if shouldPreserve {
		ok, err := m.hostdesk.PreserveSeats(ctx, party.ID, party.Size)
		if err != nil {
			m.logger.LogErr(SEAT_MANAGER, err, "failed preserve seats on processing new party")
			return nil, domain.ErrPreserveSeats
		}
		if !ok {
			m.logger.LogDebug(SEAT_MANAGER, "could not preserve seats, fallback new party to waitlist queue")
			newPartyStatus = d.PartyStatusWaiting
		}
	}

	party.Status = newPartyStatus
	if newPartyStatus == d.PartyStatusServing {
		queuedParty := &w.QueuedParty{}
		copier.Copy(queuedParty, party)
		if err := m.hostdesk.CheckIn(ctx, queuedParty); err == nil {
			m.logger.LogDebug(SEAT_MANAGER, "start serving immediately", "party", queuedParty)
			return queuedParty, nil
		}
		m.logger.LogErr(SEAT_MANAGER, err, "could not check in immediately when new party joins, fallback to waitlist queue as ready")
		party.Status = d.PartyStatusReady
	}

	queuedParty, err := m.waitlist.JoinQueue(ctx, party)
	if err != nil {
		if party.Status == d.PartyStatusReady {
			m.logger.LogErr(SEAT_MANAGER, err, "could not join waitlist as ready to serve")
			needReleaseSeats = true
		}
		return nil, domain.ErrJoinWaitlist
	}

	m.logger.LogDebug(SEAT_MANAGER, "party will join waitlist queue", "status", party.Status, "party", queuedParty)
	return queuedParty, nil
}

func (m *seatManager) checkAndAssignSeating(ctx context.Context) error {
	capacity, err := m.hostdesk.GetCurrentCapacity(ctx)
	if err != nil {
		m.logger.LogErr(SEAT_MANAGER, err, "get capacity of hostdesk failed")
		return fmt.Errorf("get capacity of hostdesk failed: %w", err)
	}

	if capacity > 0 {
		if err := m.processAvailableCapacity(ctx, capacity); err != nil {
			m.logger.LogErr(SEAT_MANAGER, err, "process available capacity")
			return err
		}
	}

	return nil
}

func (m *seatManager) processAvailableCapacity(ctx context.Context, availableSeats int) error {
	nextParty, err := m.selection.EvaluateNextParty(ctx, availableSeats)
	if err != nil {
		return fmt.Errorf("evaluate next party failed: %w", err)
	}

	if nextParty != nil {
		if err := m.hostdesk.NotifyPartyReady(ctx, nextParty); err != nil {
			return err
		}
	}
	return nil
}
