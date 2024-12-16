package config

import (
	"fmt"
	"time"

	"queue-bite/pkg/env"
)

type Config struct {
	Dev    bool
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
