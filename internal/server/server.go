package server

import (
	"fmt"
	"net/http"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	"queue-bite/internal/config"
	log "queue-bite/internal/config/logger"
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

	redis *platform.RedisComponent
}

func NewServer(
	cfg *config.Config,
	logger log.Logger,
	validate *validator.Validate,
	translators *ut.UniversalTranslator,
	cookieManager *session.CookieManager,
	cookieCfgs *config.QueueBiteCookies,
) *http.Server {
	NewServer := &Server{
		cfg:           cfg,
		logger:        logger,
		validate:      validate,
		translators:   translators,
		cookieManager: cookieManager,
		cookieCfgs:    cookieCfgs,

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
