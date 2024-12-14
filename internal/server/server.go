package server

import (
	"fmt"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"queue-bite/internal/config"
	log "queue-bite/internal/config/logger"
)

type Server struct {
	cfg    *config.Config
	logger *log.Logger
}

func NewServer(cfg *config.Config, logger *log.Logger) *http.Server {
	NewServer := &Server{
		cfg:    cfg,
		logger: logger,
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
