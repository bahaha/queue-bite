package server

import (
	"fmt"
	"net/http"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	"queue-bite/internal/config"
	log "queue-bite/internal/config/logger"
	st "queue-bite/internal/features/servicetime/service"
	"queue-bite/internal/platform"
	"queue-bite/pkg/session"
)

type Server struct {
	cfg           *config.Config
	logger        log.Logger
	validate      *validator.Validate
	translators   *ut.UniversalTranslator
	cookieManager *session.CookieManager
	cookieCfgs    *config.QueueBiteCookies

	redis                *platform.RedisComponent
	serviceTimeEstimator st.ServiceTimeEstimator
}

func NewServer(
	cfg *config.Config,
	logger log.Logger,
	serviceTimeEstimator st.ServiceTimeEstimator,
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

		redis:                platform.NewRedis(cfg, logger),
		serviceTimeEstimator: serviceTimeEstimator,
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
