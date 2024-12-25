package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	"queue-bite/internal/config"
	log "queue-bite/internal/config/logger"
	hds "queue-bite/internal/features/hostdesk/service"
	sms "queue-bite/internal/features/seatmanager/service"
	st "queue-bite/internal/features/servicetime/service"
	"queue-bite/internal/features/sse"
	wrepo "queue-bite/internal/features/waitlist/repository"
	ws "queue-bite/internal/features/waitlist/service"
	"queue-bite/internal/platform"
	eb "queue-bite/internal/platform/eventbus"
	"queue-bite/pkg/session"
)

type Server struct {
	cfg           *config.Config
	logger        log.Logger
	validate      *validator.Validate
	translators   *ut.UniversalTranslator
	cookieManager *session.CookieManager
	cookieCfgs    *config.QueueBiteCookies

	redis *platform.RedisComponent

	waitlist    ws.Waitlist
	hostdesk    hds.HostDesk
	sse         sse.ServerSentEvents
	seatmanager sms.SeatManager
}

func NewServer(
	cfg *config.Config,
	logger log.Logger,
	redis *platform.RedisComponent,
	eventRegistry *eb.EventRegistry,
	eventbus eb.EventBus,
	serviceTimeEstimator st.ServiceTimeEstimator,
	waitlistRepo wrepo.WaitlistRepositoy,
	hostdesk hds.HostDesk,
	partyProcessingStrategy sms.PartyProcessingStrategy,
	partySelectionStrategyFactory func(ws.QueuedPartyProvider) sms.PartySelectionStrategy,
) *http.Server {
	cookieManager, err := session.NewCookieManager(cfg.CookieEncryptionKey)
	if err != nil {
		logger.LogErr(log.Server, err, "cookie encryption key setup", "encryption key", cfg.CookieEncryptionKey)
	}
	localeTrans := config.NewLocaleTranslations()
	cookieCfgs := config.NewCookieConfigs(cfg)
	sseManager := sse.NewServerSentEvent(logger, eventbus)
	waitlist := ws.NewWaitlistService(logger, waitlistRepo, serviceTimeEstimator, eventbus)
	partySelection := partySelectionStrategyFactory(waitlist)
	seatManager := sms.NewSeatManager(logger, eventbus, waitlist, hostdesk, partyProcessingStrategy, partySelection, cfg.SeatManager.PreserveMaxRetries)

	NewServer := &Server{
		cfg:           cfg,
		logger:        logger,
		validate:      localeTrans.Validator,
		translators:   localeTrans.Translators,
		cookieManager: cookieManager,
		cookieCfgs:    cookieCfgs,

		waitlist:    waitlist,
		hostdesk:    hostdesk,
		sse:         sseManager,
		seatmanager: seatManager,

		redis: platform.NewRedis(cfg, logger),
	}

	NewServer.RegisterEvents(eventRegistry)
	seatManager.WatchSeatVacancy(context.Background())

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.cfg.Server.Port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	server.RegisterOnShutdown(func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		NewServer.Cleanup(ctx)
	})

	return server
}

func (s *Server) Cleanup(ctx context.Context) {
	if err := s.seatmanager.UnwatchSeatVacancy(ctx); err != nil {
		s.logger.LogErr(log.Server, err, "failed to unwatch seats vacancy")
	}
}
