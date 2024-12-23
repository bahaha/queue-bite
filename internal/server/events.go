package server

import (
	hostdesk "queue-bite/internal/features/hostdesk/domain"
	"queue-bite/internal/features/sse"
	"queue-bite/internal/platform/eventbus"
)

func (s *Server) RegisterEvents(eventRegistry *eventbus.EventRegistry) {
	eventRegistry.Register(sse.TopicNotifyPartyReady, &sse.NotifyPartyReadyEvent{})

	eventRegistry.Register(hostdesk.TopicPartyPreserved, &hostdesk.SeatsPreservedEvent{})
	eventRegistry.Register(hostdesk.TopicPartyServiceCompleted, &hostdesk.PartyServiceCompeletedEvent{})
}
