package config

import (
	"fmt"
	"time"

	"queue-bite/pkg/env"
)

type Config struct {
	Dev    bool
	Server *ServerConf
	Redis  *RedisConf
}

type ServerConf struct {
	Host               string
	Port               int
	ShutdownTimeout    time.Duration
	HealthCheckTimeout time.Duration
}

type RedisConf struct {
	Addr     string
	Password string
}

type EnvVars struct {
	Host               string        `env:"SERVER_HOST" default:"localhost"`
	Port               int           `env:"SERVER_PORT" default:"55666"`
	ShutdownTimeout    time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT_SECONDS" default:"5s"`
	HealthCheckTimeout time.Duration `env:"HEALTH_CHECK_TIMEOUT" default:"5s"`

	// Redis
	RedisHost     string `env:"WAITLIST_REDIS_HOST"`
	RedisPort     int    `env:"WAITLIST_REDIS_PORT" default:"6379"`
	RedisPassword string `env:"WAITLIST_REDIS_PASSWORD"`
}

func LoadEnvConfig(getenv func(string) string) (*EnvVars, error) {
	loader := env.NewEnvLoader(env.WithEnvSource(getenv))
	cfg := &EnvVars{}
	err := loader.Parse(cfg)
	return cfg, err
}

func NewConfig(getenv func(string) string) (*Config, error) {
	cfg, err := LoadEnvConfig(getenv)
	if err != nil {
		return nil, fmt.Errorf("Invalid server configuration, check your environment values: %v", err)
	}

	return &Config{
		Dev: getenv("APP_ENV") != "production",
		Server: &ServerConf{
			Host:               cfg.Host,
			Port:               cfg.Port,
			ShutdownTimeout:    cfg.ShutdownTimeout,
			HealthCheckTimeout: cfg.HealthCheckTimeout,
		},
		Redis: &RedisConf{
			Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
			Password: cfg.RedisPassword,
		},
	}, nil
}
