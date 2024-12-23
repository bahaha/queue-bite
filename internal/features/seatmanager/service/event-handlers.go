package service

import (
	"context"
	hdd "queue-bite/internal/features/hostdesk/domain"
	"queue-bite/internal/platform/eventbus"
)

func (m *seatManager) handleSeatPreservedEvent(ctx context.Context, event eventbus.Event) error {
	e := event.(*hdd.SeatsPreservedEvent)

	if err := m.waitlist.HandlePartyReady(ctx, e.PartyID); err != nil {
		m.logger.LogErr(SEAT_MANAGER, err, "failed to make party ready", "event", e)
		return err
	}

	return nil
}

func (m *seatManager) handlePartyServiceCompleted(ctx context.Context, event eventbus.Event) error {
	m.checkAndAssignSeating(ctx)
	return nil
}
