package handler

import (
	"net/http"

	"queue-bite/pkg/session"

	log "queue-bite/internal/config/logger"
	"queue-bite/internal/features/seatmanager/handler/view"
	w "queue-bite/internal/features/waitlist/domain"
	ws "queue-bite/internal/features/waitlist/service"

	"github.com/a-h/templ"
)

type VitrineHandler struct{}

func NewVitrineHandler() *VitrineHandler {
	return &VitrineHandler{}
}

var VITRINE = "seatmanager/vitrine"

func (h *VitrineHandler) HandleVitrineDisplay(
	logger log.Logger,
	cookieManager *session.CookieManager,
	cookieQueuedParty *session.CookieConfig,
	waitlist ws.Waitlist,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		status, err := waitlist.GetQueueStatus(r.Context())
		if err != nil {
			logger.LogErr(VITRINE, err, "failed to fetch queue status")
			h.renderVisitorView(w, r, status)
			return
		}

		var partySession PartySession
		if err := cookieManager.GetCookie(r, cookieQueuedParty, &partySession); err != nil {
			h.renderVisitorView(w, r, status)
			return
		}

		queuedParty, err := waitlist.GetQueuedParty(r.Context(), partySession.PartyID)
		if err != nil || queuedParty == nil {
			logger.LogDebug("party no longer in queue, clearing cookie",
				"party_id", partySession.PartyID)
			cookieManager.ClearCookie(w, cookieQueuedParty)
			h.renderVisitorView(w, r, status)
			return
		}

		logger.LogDebug(VITRINE, "rendering queued party view", "party_id", queuedParty.ID, "position", queuedParty.Position)
		h.renderQueuedPartyView(w, r, queuedParty, status)
	}
}

func (h *VitrineHandler) renderVisitorView(
	w http.ResponseWriter,
	r *http.Request,
	status *w.QueueStatus,
) {
	props := view.ToVitrineProps(nil, status)
	templ.Handler(view.VitrinePage(props)).ServeHTTP(w, r)
}

func (h *VitrineHandler) renderQueuedPartyView(
	w http.ResponseWriter,
	r *http.Request,
	party *w.QueuedParty,
	status *w.QueueStatus,
) {
	props := view.ToVitrineProps(party, status)
	templ.Handler(view.VitrinePage(props)).ServeHTTP(w, r)
}