package service

import (
	"context"
	"queue-bite/internal/features/hostdesk/domain"
)

type HostDesk interface {
	GetCurrentCapacity(ctx context.Context) (int, error)

	NotifyPartyReady(ctx context.Context, partyID string) error

	GetPartyServiceState(ctx context.Context, partyID string) (*domain.PartyServiceState, error)
}
