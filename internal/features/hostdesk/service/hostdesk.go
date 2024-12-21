package service

import (
	"context"

	d "queue-bite/internal/domain"
	w "queue-bite/internal/features/waitlist/domain"
)

type HostDesk interface {
	GetCurrentCapacity(ctx context.Context) (int, error)

	NotifyPartyReady(ctx context.Context, party *w.QueuedParty) error

	PreserveSeats(ctx context.Context, partyID d.PartyID, seats int) (bool, error)
}
