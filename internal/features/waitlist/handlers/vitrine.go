package waitlist

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/jinzhu/copier"

	"queue-bite/internal/config"
	log "queue-bite/internal/config/logger"
	"queue-bite/internal/features/waitlist/services"
	view "queue-bite/internal/features/waitlist/views"
	"queue-bite/pkg/session"
)

type waitlistVitrineHandler struct {
	service services.WaitlistService
}

func newWaitlistVitrineHandler(service services.WaitlistService) *waitlistVitrineHandler {
	return &waitlistVitrineHandler{
		service: service,
	}
}

func (h *waitlistVitrineHandler) GetVitrineDisplay(
	logger log.Logger,
	cookieManager *session.CookieManager,
	cookieCfgs *config.QueueBiteCookies,
) http.HandlerFunc {
	cookieCfg := &cookieCfgs.QueuedPartyCookie

	return func(w http.ResponseWriter, r *http.Request) {
		clientPartyInfo := &services.QueuedParty{}
		if err := cookieManager.GetCookie(r, &cookieCfgs.QueuedPartyCookie, clientPartyInfo); err != nil {
			logger.LogErr(WAITLIST, err, "party is not in queued")
			templ.Handler(vitrineViewForVisitor()).ServeHTTP(w, r)
			return
		}

		queued, err := h.service.GetQueuedParty(r.Context(), clientPartyInfo.ID)
		if queued == nil || err != nil {
			logger.LogErr(WAITLIST, err, "unprocess entity by session-cookie party ID", "ID", clientPartyInfo.ID)
			cookieManager.ClearCookie(w, cookieCfg)
			templ.Handler(vitrineViewForVisitor()).ServeHTTP(w, r)
			return
		}

		logger.LogDebug(WAITLIST, "queued party re-enter the waitlist", "queued party", queued)
		templ.Handler(vitrineViewForQueuedParty(queued)).ServeHTTP(w, r)
	}
}

func vitrineViewForVisitor() templ.Component {
	return view.Vitrine(view.NewVitrinePageDataForVisitor(view.NewJoinFormData()))
}

func vitrineViewForQueuedParty(queued *services.QueuedParty) templ.Component {
	props := &view.QueuedPartyProps{}
	copier.Copy(props, queued)

	return view.Vitrine(view.NewVitrinePageDataForQueuedParty(props))
}
