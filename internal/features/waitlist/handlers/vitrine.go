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
		status, err := h.service.GetQueueStatus(r.Context())
		if err != nil {
			logger.LogErr(WAITLIST, err, "could not fetch queue status from waitlist")
			templ.Handler(vitrineViewForVisitor(nil)).ServeHTTP(w, r)
			return
		}

		clientPartyInfo := &services.QueuedParty{}
		if err := cookieManager.GetCookie(r, &cookieCfgs.QueuedPartyCookie, clientPartyInfo); err != nil {
			logger.LogErr(WAITLIST, err, "party is not in queued")
			templ.Handler(vitrineViewForVisitor(status)).ServeHTTP(w, r)
			return
		}

		queued, err := h.service.GetQueuedParty(r.Context(), clientPartyInfo.ID)
		if queued == nil || err != nil {
			logger.LogErr(WAITLIST, err, "unprocess entity by session-cookie party ID", "ID", clientPartyInfo.ID)
			cookieManager.ClearCookie(w, cookieCfg)
			templ.Handler(vitrineViewForVisitor(status)).ServeHTTP(w, r)
			return
		}

		logger.LogDebug(WAITLIST, "queued party re-enter the waitlist", "queued party", queued)
		templ.Handler(vitrineViewForQueuedParty(queued, status)).ServeHTTP(w, r)
	}
}

func vitrineViewForVisitor(queueStatus *services.QueueStatus) templ.Component {
	pageProps := &view.VitrinePageData{
		Form: view.NewJoinFormData(),
	}

	if queueStatus != nil {
		statusProps := &view.QueueStatusProps{}
		copier.Copy(statusProps, queueStatus)
		pageProps.QueueStatus = statusProps
	}

	return view.Vitrine(pageProps)
}

func vitrineViewForQueuedParty(queued *services.QueuedParty, queueStatus *services.QueueStatus) templ.Component {
	props := &view.QueuedPartyProps{}
	copier.Copy(props, queued)

	pageProps := &view.VitrinePageData{
		QueuedParty: props,
	}

	if queueStatus != nil {
		statusProps := &view.QueueStatusProps{
			TotalParties:  queueStatus.TotalParties,
			EstimatedWait: queued.EstimatedWait,
		}
		pageProps.QueueStatus = statusProps
	}

	return view.Vitrine(pageProps)
}
