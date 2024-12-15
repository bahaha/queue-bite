package config

import (
	"strconv"
	"time"
)

type Config struct {
	Dev    bool
	Server *ServerConfig
	Redis  *RedisConfig
}

type ServerConfig struct {
	Host               string
	Port               int
	ShutdownTimeout    time.Duration
	HealthCheckTimeout time.Duration
}

func NewConfig(getenv func(string) string) (*Config, error) {

	port, _ := strconv.Atoi(getenvOrDefault(getenv, "SERVER_PORT", "55666"))
	shutdownTimeout, _ := strconv.Atoi(getenvOrDefault(getenv, "SERVER_SHUTDOWN_TIMEOUT_SECONDS", "5"))
	healthcheckTimeout, _ := strconv.Atoi(getenvOrDefault(getenv, "HEALTH_CHECK_TIMEOUT", "5"))

	return &Config{
		Dev: getenv("APP_ENV") != "production",
		Server: &ServerConfig{
			Host:               getenvOrDefault(getenv, "SERVER_HOST", "localhost"),
			Port:               port,
			ShutdownTimeout:    time.Duration(shutdownTimeout) * time.Second,
			HealthCheckTimeout: time.Duration(healthcheckTimeout) * time.Second,
		},
		Redis: NewRedisConfig(getenv),
	}, nil
}

func getenvOrDefault(getenv func(string) string, key, defaultValue string) string {
	if value := getenv(key); value != "" {
		return value
	}
	return defaultValue
}
