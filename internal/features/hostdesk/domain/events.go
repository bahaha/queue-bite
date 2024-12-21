package domain

import (
	d "queue-bite/internal/domain"
	"queue-bite/internal/platform/eventbus"
)

type SeatsPreservedEvent struct {
	PartyID d.PartyID
}

func (e SeatsPreservedEvent) Topic() string {
	return TopicPartyPreserved
}

func (e SeatsPreservedEvent) NewEvent() eventbus.Event {
	return &SeatsPreservedEvent{}
}

const (
	TopicPartyPreserved = "hd.party.preserved"
)
