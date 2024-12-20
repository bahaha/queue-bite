package server

import (
	"fmt"
	"net/http"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	"queue-bite/internal/config"
	log "queue-bite/internal/config/logger"
	hds "queue-bite/internal/features/hostdesk/service"
	sms "queue-bite/internal/features/seatmanager/service"
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

	waitlist ws.Waitlist
}

func NewServer(
	cfg *config.Config,
	logger log.Logger,
	redis *platform.RedisComponent,
	eventRegistry *eb.EventRegistry,
	eventbus eb.EventBus,
	waitlist ws.Waitlist,
	hostdesk hds.HostDesk,
	seatmanager sms.SeatManager,
) *http.Server {
	cookieManager, err := session.NewCookieManager(cfg.CookieEncryptionKey)
	if err != nil {
		logger.LogErr(log.Server, err, "cookie encryption key setup", "encryption key", cfg.CookieEncryptionKey)
	}
	localeTrans := config.NewLocaleTranslations()
	cookieCfgs := config.NewCookieConfigs(cfg)

	NewServer := &Server{
		cfg:           cfg,
		logger:        logger,
		validate:      localeTrans.Validator,
		translators:   localeTrans.Translators,
		cookieManager: cookieManager,
		cookieCfgs:    cookieCfgs,

		waitlist: waitlist,

		redis: platform.NewRedis(cfg, logger),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.cfg.Server.Port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
