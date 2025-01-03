package service

import (
	"context"
	"queue-bite/internal/domain"
	w "queue-bite/internal/features/waitlist/domain"
)

type SeatingContext struct {
	SeatsAvailable bool
	*w.QueueStatus
}

type PartyProcessingStrategy interface {
	DeterminePartyState(ctx context.Context, seatingCtx *SeatingContext) (domain.PartyStatus, bool)
}

type instantServiceProcessor struct{}

func NewInstantServingStrategy() PartyProcessingStrategy {
	return &instantServiceProcessor{}
}

func (p *instantServiceProcessor) DeterminePartyState(ctx context.Context, seatingCtx *SeatingContext) (domain.PartyStatus, bool) {
	if seatingCtx.TotalParties != 0 || !seatingCtx.SeatsAvailable {
		return domain.PartyStatusWaiting, false
	}
	return domain.PartyStatusServing, true
}

type fairOrderProcessor struct{}

func NewFairOrderStrategy() PartyProcessingStrategy {
	return &fairOrderProcessor{}
}

func (p *fairOrderProcessor) DeterminePartyState(ctx context.Context, seatingCtx *SeatingContext) (domain.PartyStatus, bool) {
	if !seatingCtx.SeatsAvailable || seatingCtx.WaitingParties != 0 {
		return domain.PartyStatusWaiting, false
	}

	return domain.PartyStatusReady, true
}
