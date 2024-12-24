package config

import (
	"fmt"
	"time"

	"queue-bite/pkg/env"
)

type Config struct {
	Dev bool
	// TODO: validate encryption key length by adding functionality with github.com/go-playground/validator/v10 in env module
	CookieEncryptionKey string `env:"SECRET_COOKIE_ENCRYPTION_KEY" required:"T"`

	Server struct {
		Host               string        `env:"SERVER_HOST" default:"localhost"`
		Port               int           `env:"SERVER_PORT" default:"55666"`
		ShutdownTimeout    time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT_SECONDS" default:"5s"`
		HealthCheckTimeout time.Duration `env:"HEALTH_CHECK_TIMEOUT" default:"5s"`
	}

	Redis struct {
		Addr     string
		Host     string `env:"WAITLIST_REDIS_HOST"`
		Port     int    `env:"WAITLIST_REDIS_PORT" default:"6379"`
		Password string `env:"WAITLIST_REDIS_PASSWORD"`
	}
	Waitlist struct {
		ScanChunkSize int           `env:"WAITLIST_SCAN_CHUNK_SIZE" default:"5"`
		EntityTTL     time.Duration `env:"WAITLIST_ENTITY_TTL" default:"24h"`
	}
	ServiceEstimator struct {
		FixedRateUnit time.Duration `env:"FIXED_RATE_SERVICE_ESTIMATOR_UNIT" default:"3s"`
	}
	HostDesk struct {
		InstantServeHostDeskSeatCapacity   int           `env:"INSTANT_SERVE_HOST_DESK_SEAT_CAPACITY" default:"10"`
		LinearServiceTimerDurationPerGuest time.Duration `env:"LINEAR_SERVICE_TIMER_DURATION_PER_GUEST" default:"3s"`
	}
	SeatManager struct {
		PreserveMaxRetries int `env:"PRESERVE_SEAT_MAX_RETRIES" default:"3"`
	}
}

func LoadEnvConfig(getenv func(string) string) (*Config, error) {
	loader := env.NewEnvLoader(env.WithEnvSource(getenv))
	cfg := &Config{}
	err := loader.Parse(cfg)
	return cfg, err
}

func NewConfig(getenv func(string) string) (*Config, error) {
	cfg, err := LoadEnvConfig(getenv)
	if err != nil {
		return nil, fmt.Errorf("Invalid server configuration, check your environment values: %v", err)
	}

	cfg.Dev = getenv("APP_ENV") != "production"
	cfg.Redis.Addr = fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	return cfg, nil
}
