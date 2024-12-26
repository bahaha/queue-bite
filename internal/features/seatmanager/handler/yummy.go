package handler

import (
	"net/http"
	log "queue-bite/internal/config/logger"
	hd "queue-bite/internal/features/hostdesk/service"
	"queue-bite/internal/features/seatmanager/domain"
	"queue-bite/internal/features/seatmanager/handler/view"
	"queue-bite/pkg/session"

	"github.com/a-h/templ"
)

func (h *seatManagerHandler) HandleServingDisplay(
	logger log.Logger,
	cookieManager *session.CookieManager,
	cookieQueuedParty *session.CookieConfig,
	hostdesk hd.HostDesk,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var partySession domain.PartySession
		if err := cookieManager.GetCookie(r, cookieQueuedParty, &partySession); err != nil {
			logger.LogDebug(SEAT_MANAGER_CHECKIN, "could not access session cookie from yummy")
			redirectToVisitPage(w, r)
			return
		}

		exists := hostdesk.HasPartyOccupiedSeat(r.Context(), partySession.ID)
		if !exists {
			redirectToVisitPage(w, r)
			return
		}
		templ.Handler(view.Yummy(view.NewYummyProps(&partySession))).ServeHTTP(w, r)
	}
}
