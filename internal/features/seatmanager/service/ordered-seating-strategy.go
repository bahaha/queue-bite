package service

import (
	"context"

	w "queue-bite/internal/features/waitlist/domain"
	ws "queue-bite/internal/features/waitlist/service"
)

type OrderedSeatingStrategy struct {
	waitlist ws.QueuedPartyProvider
}

func NewOrderedSeatingStrategy(waitlist ws.QueuedPartyProvider) SeatingStrategy {
	return &OrderedSeatingStrategy{
		waitlist: waitlist,
	}
}

func (s *OrderedSeatingStrategy) EvaluateNextParty(ctx context.Context, vacancySeats int) (*w.QueuedParty, error) {
	queuedParties, err := s.waitlist.GetQueuedParties(ctx)
	if err != nil {
		return nil, err
	}

	for party := range queuedParties {
		if party.Size < vacancySeats {
			return party, nil
		}
	}

	return nil, nil
}
