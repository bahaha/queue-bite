package service

import (
	"context"

	waitlist "queue-bite/internal/features/waitlist/domain"
	waitlist_svc "queue-bite/internal/features/waitlist/service"
)

type OrderedSeatingStrategy struct {
	waitlist waitlist_svc.Waitlist
}

func NewOrderedSeatingStrategy(waitlist waitlist_svc.Waitlist) *OrderedSeatingStrategy {
	return &OrderedSeatingStrategy{
		waitlist: waitlist,
	}
}

func (s *OrderedSeatingStrategy) EvaluateNextParty(ctx context.Context, vacancySeats int) (*waitlist.QueuedParty, error) {
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
