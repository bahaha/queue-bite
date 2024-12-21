package domain

import (
	"queue-bite/internal/platform/eventbus"
)

type EvaluateNextPartyEvent struct {
}

func (e EvaluateNextPartyEvent) Topic() string {
	return TopicEvaluateNextParty
}

func (e EvaluateNextPartyEvent) NewEvent() eventbus.Event {
	return &EvaluateNextPartyEvent{}
}

const (
	TopicEvaluateNextParty = "sm:evaluate"
)
