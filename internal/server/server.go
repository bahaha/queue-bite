package server

import (
	"fmt"
	"net/http"
	"time"

	"queue-bite/internal/config"
	log "queue-bite/internal/config/logger"
	"queue-bite/internal/platform"
)

type Server struct {
	cfg    *config.Config
	logger log.Logger

	redis *platform.RedisComponent
}

func NewServer(cfg *config.Config, logger log.Logger) *http.Server {
	NewServer := &Server{
		cfg:    cfg,
		logger: logger,

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
