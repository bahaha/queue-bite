package waitlist

import (
	"net/http"

	"github.com/a-h/templ"

	"queue-bite/internal/config"
	log "queue-bite/internal/config/logger"
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
		queuedParty := &QueuedParty{}
		if err := cookieManager.GetCookie(r, &cookieCfgs.QueuedPartyCookie, queuedParty); err != nil {
			templ.Handler(view.Vitrine(&view.VitrinePageData{
				Form: view.NewJoinFormData(),
			})).ServeHTTP(w, r)
			return
		}

		logger.LogDebug(FEAT_WAITLIST, "found queued party from cookie", "queued party", queuedParty)
		templ.Handler(view.Vitrine(&view.VitrinePageData{
			QueueEntry: &view.QueueEntry{
				QueueOrder: 5,
				PartyName:  queuedParty.Name,
			},
		})).ServeHTTP(w, r)
	}
}
