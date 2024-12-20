package domain

import (
	d "queue-bite/internal/domain"
	"queue-bite/internal/platform/eventbus"
	"time"
)

type PartyReadyEvent struct {
	PartyID d.PartyID
	ReadyAt time.Time
}

func (e PartyReadyEvent) Topic() string {
	return TopicPartyReady
}

func (e PartyReadyEvent) NewEvent() eventbus.Event {
	return &PartyReadyEvent{}
}

const (
	TopicPartyReady = "hd.party.ready"
)
