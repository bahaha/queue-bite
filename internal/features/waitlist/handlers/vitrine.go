package waitlist

import (
	"net/http"

	"github.com/a-h/templ"

	"queue-bite/internal/config"
	log "queue-bite/internal/config/logger"
	"queue-bite/internal/features/waitlist/services"
	view "queue-bite/internal/features/waitlist/views"
	"queue-bite/pkg/session"
)

type waitlistVitrineHandler struct{}

func newWaitlistVitrineHandler() *waitlistVitrineHandler {
	return &waitlistVitrineHandler{}
}

func (h *waitlistVitrineHandler) GetVitrineDisplay(
	logger log.Logger,
	cookieManager *session.CookieManager,
	cookieCfgs *config.QueueBiteCookies,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		queuedParty := &services.QueuedParty{}
		if err := cookieManager.GetCookie(r, &cookieCfgs.QueuedPartyCookie, queuedParty); err != nil {
			logger.LogErr(WAITLIST, err, "party is not in queued")
			templ.Handler(view.Vitrine(&view.VitrinePageData{
				Form: view.NewJoinFormData(),
			})).ServeHTTP(w, r)
			return
		}

		logger.LogDebug(WAITLIST, "found queued party from cookie", "queued party ID", queuedParty.ID)
		templ.Handler(view.Vitrine(&view.VitrinePageData{
			QueuedPartyProps: &view.QueuedPartyProps{},
		})).ServeHTTP(w, r)
	}
}
