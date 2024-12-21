package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	log "queue-bite/internal/config/logger"
	"queue-bite/internal/domain"
	sse "queue-bite/internal/features/sse"

	waitlist "queue-bite/internal/features/waitlist/service"
)

func HandleQueuedPartyServerSentEventConn(
	logger log.Logger,
	sse sse.ServerSentEvents,
	waitlist waitlist.Waitlist,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "partyID")
		partyID := domain.PartyID(id)

		if !waitlist.HasPartyExists(r.Context(), partyID) {
			logger.LogDebug("sse/conn", "could not find party in waitlist queue", "party_id", partyID)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		sse.RegisterClient(w, partyID)
		defer sse.UnregisterClient(partyID)

		<-r.Context().Done()
		logger.LogDebug("sse/conn", "party server sent event disconnected", "party_id", partyID)
	}
}

func ManuallyEvaluateNextQueuedParty(s sse.ServerSentEvents) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "partyID")
		s.HandleNotifyPartyReady(r.Context(), &sse.NotifyPartyReadyEvent{PartyID: domain.PartyID(id)})
	}
}
