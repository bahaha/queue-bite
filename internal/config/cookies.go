package config

import (
	"queue-bite/pkg/session"
	"time"
)

type QueueBiteCookies struct {
	QueuedPartyCookie session.CookieConfig
}

func NewCookieConfigs(cfg *Config) *QueueBiteCookies {
	return &QueueBiteCookies{
		QueuedPartyCookie: *session.
			NewCookieConfig("qb_qp", cfg.Server.Host).
			WithHttpOnly(true).
			WithSecure(!cfg.Dev).
			WithTTL(12 * time.Hour),
	}
}
