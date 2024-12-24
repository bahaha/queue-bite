package handler

import (
	"net/http"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/seatmanager/service"
	w "queue-bite/internal/features/waitlist/domain"
	"queue-bite/pkg/session"
)

var SEAT_MANAGER_CHECKIN = "seatmanager/check-in"

func (h *seatManagerHandler) HandlePartyCheckIn(
	logger log.Logger,
	seatManager service.SeatManager,
	cookieManager *session.CookieManager,
	cookieQueuedParty *session.CookieConfig,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var partySession PartySession
		if err := cookieManager.GetCookie(r, cookieQueuedParty, &partySession); err != nil {
			logger.LogDebug(SEAT_MANAGER_CHECKIN, "could not access session cookie from check-in")
			redirectToVisitPage(w, r)
			return
		}

		err := seatManager.PartyCheckIn(r.Context(), d.PartyID(partySession.PartyID))
		if err != nil {
			handleErrorOnCheckIn(logger, w, r, cookieManager, cookieQueuedParty, err)
			return
		}

		logger.LogDebug(SEAT_MANAGER_CHECKIN, "party has just checked-in", "party id", partySession.PartyID)
		w.Header().Add("HX-Location", "/yummy")
	}
}

func handleErrorOnCheckIn(
	logger log.Logger,
	resp http.ResponseWriter,
	req *http.Request,
	cookieManager *session.CookieManager,
	cookieQueuedParty *session.CookieConfig,
	err error,
) {
	logger.LogErr(SEAT_MANAGER_CHECKIN, err, "handle ready party check-in failed")
	switch err {
	case w.ErrPartyNotFound:
		cookieManager.ClearCookie(resp, cookieQueuedParty)
		redirectToVisitPage(resp, req)
		return
	}

	http.Error(resp, "Failed to check-in", http.StatusInternalServerError)
}
