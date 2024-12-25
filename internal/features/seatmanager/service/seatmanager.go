package service

import (
	"context"
	"errors"
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
	PartyCheckIn(ctx context.Context, partyID d.PartyID) error
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

	preserveMaxRetries int
}

func NewSeatManager(
	logger log.Logger,
	eventbus eventbus.EventBus,
	waitlist waitlist.Waitlist,
	hostdesk hostdesk.HostDesk,
	processing PartyProcessingStrategy,
	selection PartySelectionStrategy,
	preserveMaxRetries int,
) SeatManager {
	return &seatManager{
		logger:             logger,
		eventbus:           eventbus,
		waitlist:           waitlist,
		hostdesk:           hostdesk,
		processing:         processing,
		selection:          selection,
		preserveMaxRetries: preserveMaxRetries,
	}
}

func (m *seatManager) WatchSeatVacancy(ctx context.Context) error {
	m.eventbus.Subscribe(hdd.TopicPartyPreserved, m.handleSeatPreservedEvent)
	m.eventbus.Subscribe(hdd.TopicPartyServiceCompleted, m.handlePartyServiceCompleted)
	return nil
}

func (m *seatManager) UnwatchSeatVacancy(ctx context.Context) error {
	m.logger.LogDebug(SEAT_MANAGER, "Seat manager stop observing")
	return nil
}

func (m *seatManager) ProcessNewParty(ctx context.Context, party *d.Party) (*w.QueuedParty, error) {
	for retries := 0; retries < m.preserveMaxRetries; retries++ {
		capacity, version, err := m.hostdesk.GetCurrentCapacity(ctx)
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
			ok, err := m.hostdesk.PreserveSeats(ctx, party.ID, party.Size, version)
			if err != nil {
				m.logger.LogErr(SEAT_MANAGER, err, "failed preserve seats on processing new party", "retry", retries)
				if errors.Is(err, d.ErrVersionMismatch) {
					continue
				}
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

	return nil, d.ErrTooManyOptimisticLockRetries
}

func (m *seatManager) PartyCheckIn(ctx context.Context, partyID d.PartyID) error {
	party, err := m.waitlist.GetQueuedParty(ctx, partyID)
	if err != nil {
		return err
	}

	// TODO: transaction like leave queue, but checkin fails
	if party.Status != d.PartyStatusServing {
		if err := m.waitlist.LeaveQueue(ctx, party.ID); err != nil {
			m.logger.LogErr(SEAT_MANAGER, err, "leave queue failed", "party", party)
			return err
		}
		m.logger.LogDebug(SEAT_MANAGER, "party leave queue by checking in", "party", party)
	}

	if err := m.hostdesk.CheckIn(ctx, party); err != nil {
		m.logger.LogErr(SEAT_MANAGER, err, "check in failed", "party", party)
		return err
	}
	m.logger.LogDebug(SEAT_MANAGER, "party check in", "party", party)
	return nil
}

func (m *seatManager) checkAndAssignSeating(ctx context.Context) error {
	capacity, _, err := m.hostdesk.GetCurrentCapacity(ctx)
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
		m.logger.LogDebug(SEAT_MANAGER, "available seats are enough for next party", "party", nextParty)
		if err := m.hostdesk.NotifyPartyReady(ctx, nextParty); err != nil {
			return err
		}
	}
	return nil
}
