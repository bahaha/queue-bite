package sse

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a-h/templ"

	d "queue-bite/internal/domain"
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

	notifyClient(client, TopicNotifyPartyReady, renderSeatReady(e.PartyID))
	s.logger.LogDebug(SSE, "write seat ready button for next party", "party id", e.PartyID)
	return nil
}

func renderSeatReady(partyID d.PartyID) templ.Component {
	return view.QueueStatusView(view.NewReadyPartyProps(partyID))
}

func notifyClient(client *Client, eventName string, comp templ.Component) {
	fmt.Fprintf(client.Writer, "event: %s\n", eventName)
	fmt.Fprintf(client.Writer, "data: ")
	comp.Render(context.Background(), client.Writer)
	fmt.Fprintf(client.Writer, "\n\n")
	client.Writer.(http.Flusher).Flush()
}
