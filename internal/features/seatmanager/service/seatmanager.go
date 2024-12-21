package service

import (
	"context"
	"fmt"

	log "queue-bite/internal/config/logger"
	hdd "queue-bite/internal/features/hostdesk/domain"
	hostdesk "queue-bite/internal/features/hostdesk/service"
	w "queue-bite/internal/features/waitlist/domain"
	waitlist "queue-bite/internal/features/waitlist/service"
	"queue-bite/internal/platform/eventbus"
)

var SEAT_MANAGER = "seatmanager"

type SeatManager interface {
	WatchSeatVacancy(ctx context.Context) error
	UnwatchSeatVacancy(ctx context.Context) error
}

type SeatingStrategy interface {
	EvaluateNextParty(ctx context.Context, vacancySeats int) (*w.QueuedParty, error)
}

type seatManager struct {
	logger   log.Logger
	eventbus eventbus.EventBus
	waitlist waitlist.Waitlist
	hostdesk hostdesk.HostDesk
	strategy SeatingStrategy

	// channel for cleanup
	stopCh chan struct{}
}

func NewSeatManager(
	logger log.Logger,
	eventbus eventbus.EventBus,
	waitlist waitlist.Waitlist,
	hostdesk hostdesk.HostDesk,
	strategy SeatingStrategy,
) *seatManager {
	return &seatManager{
		logger:   logger,
		eventbus: eventbus,
		waitlist: waitlist,
		hostdesk: hostdesk,
		strategy: strategy,
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
	nextParty, err := m.strategy.EvaluateNextParty(ctx, availableSeats)
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
