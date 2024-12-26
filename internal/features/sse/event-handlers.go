package sse

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a-h/templ"

	"queue-bite/internal/features/seatmanager/handler/view"
	"queue-bite/internal/platform/eventbus"
)

func (s *sse) HandleNotifyPartyReady(ctx context.Context, event eventbus.Event) error {
	e := event.(*NotifyPartyReadyEvent)
	client := s.getClient(e.PartyID)
	if client == nil {
		s.logger.LogDebug(SSE, "no registered client found on this server", "party id", e.PartyID)
		return nil
	}

	notifyClient(client, TopicNotifyPartyReady, view.QueueStatusView(view.NewReadyPartyProps(e.PartyID)))
	s.logger.LogDebug(SSE, "write seat ready button for next party", "party id", e.PartyID)
	return nil
}

func (s *sse) HandleNotifyPartyQueueStatusUpdate(ctx context.Context, event eventbus.Event) error {
	e := event.(*NotifyPartyQueueStatusUpdateEvent)
	client := s.getClient(e.QueuedParty.ID)
	if client == nil {
		s.logger.LogDebug(SSE, "no registered client found on this server", "party id", e.QueuedParty.ID)
		return nil
	}

	notifyClient(client, TopicNotifyPartyQueueStatusUpdate, view.QueueStatusView(view.NewQueuedPartyProps(e.QueuedParty)))
	s.logger.LogDebug(SSE, "update queue status for waiting party", "party id", e.QueuedParty.ID)
	return nil
}

func notifyClient(client *Client, eventName string, comp templ.Component) {
	fmt.Fprintf(client.Writer, "event: %s\n", eventName)
	fmt.Fprintf(client.Writer, "data: ")
	comp.Render(context.Background(), client.Writer)
	fmt.Fprintf(client.Writer, "\n\n")
	client.Writer.(http.Flusher).Flush()
}
