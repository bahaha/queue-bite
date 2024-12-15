package config

import (
	"fmt"
	"strconv"
)

type RedisConfig struct {
	host string
	port int

	Password string
}

func NewRedisConfig(getenv func(string) string) *RedisConfig {
	port, _ := strconv.Atoi(getenvOrDefault(getenv, "WAITLIST_REDIS_PORT", "6379"))

	return &RedisConfig{
		host: getenvOrDefault(getenv, "WAITLIST_REDIS_HOST", "localhost"),
		port: port,
	}
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.host, r.port)
}
