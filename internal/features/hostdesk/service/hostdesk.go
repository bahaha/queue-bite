package service

import (
	"context"

	d "queue-bite/internal/domain"
	"queue-bite/internal/features/hostdesk/domain"
)

type HostDesk interface {
	GetCurrentCapacity(ctx context.Context) (int, error)

	NotifyPartyReady(ctx context.Context, partyID d.PartyID) error

	GetPartyServiceState(ctx context.Context, partyID d.PartyID) (*domain.PartyServiceState, error)
}
