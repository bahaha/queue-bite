package sse

import (
	d "queue-bite/internal/domain"
	wld "queue-bite/internal/features/waitlist/domain"
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
	TopicNotifyPartyReady             = "notify:party:ready"
	TopicNotifyPartyQueueStatusUpdate = "notify:party:queue_update"
)

type NotifyPartyQueueStatusUpdateEvent struct {
	QueuedParty *wld.QueuedParty
}

func (e NotifyPartyQueueStatusUpdateEvent) Topic() string {
	return TopicNotifyPartyQueueStatusUpdate
}

func (e NotifyPartyQueueStatusUpdateEvent) NewEvent() eventbus.Event {
	return &NotifyPartyQueueStatusUpdateEvent{}
}
