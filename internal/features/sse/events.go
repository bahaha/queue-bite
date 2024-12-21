package sse

import (
	d "queue-bite/internal/domain"
	"queue-bite/internal/platform/eventbus"
)

type NotifyPartyReadyEvent struct {
	PartyID d.PartyID
}

func (e NotifyPartyReadyEvent) Topic() string {
	return TopicNotifyPartyReady
}

func (e NotifyPartyReadyEvent) NewEvent() eventbus.Event {
	return &NotifyPartyReadyEvent{}
}

const (
	TopicNotifyPartyReady = "notify:party:ready"
)
