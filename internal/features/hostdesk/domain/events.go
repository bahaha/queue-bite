package domain

import (
	d "queue-bite/internal/domain"
	"queue-bite/internal/platform/eventbus"
)

type SeatsPreservedEvent struct{ PartyID d.PartyID }

func (e SeatsPreservedEvent) Topic() string { return TopicPartyPreserved }

func (e SeatsPreservedEvent) NewEvent() eventbus.Event {
	return &SeatsPreservedEvent{}
}

type PartyServiceCompeletedEvent struct{ PartyID d.PartyID }

func (e PartyServiceCompeletedEvent) Topic() string { return TopicPartyServiceCompleted }

func (e PartyServiceCompeletedEvent) NewEvent() eventbus.Event {
	return &PartyServiceCompeletedEvent{}
}

const (
	TopicPartyPreserved        = "hd.party.preserved"
	TopicPartyServiceCompleted = "hd.party.serviced"
)
