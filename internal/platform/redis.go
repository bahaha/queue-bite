package platform

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"

	"queue-bite/internal/config"
	"queue-bite/internal/config/logger"
)

type SystemComponents interface {
	Name() string
	Health() map[string]string
}

type RedisComponent struct {
	cfg    *config.Config
	logger log.Logger
	Client *redis.Client
}

func NewRedis(cfg *config.Config, logger log.Logger) *RedisComponent {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	})

	return &RedisComponent{cfg: cfg, logger: logger, Client: client}
}

func (s *RedisComponent) Name() string {
	return "redis"
}

func (s *RedisComponent) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.HealthCheckTimeout)
	defer cancel()

	stats := make(map[string]string)
	stats["status"] = "up"

	_, err := s.Client.Ping(ctx).Result()
	if err != nil {
		s.logger.LogErr(log.Redis, err, "redis server down")
		stats["status"] = "down"
	}

	info, err := s.Client.Info(ctx).Result()
	if err != nil {
		s.logger.LogErr(log.Redis, err, "failed to retrieve Redis info")
		stats["message"] = fmt.Sprintf("failed to retrieve Redis info: %v", err)
		return stats
	}

	redisInfo := parseRedisInfo(info)
	stats["version"] = redisInfo["redis_version"]
	stats["connected_clients"] = redisInfo["connected_clients"]
	stats["used_memory"] = redisInfo["used_memory"]
	stats["used_memory_peak"] = redisInfo["used_memory_peak"]
	stats["uptime_seconds"] = redisInfo["uptime_in_seconds"]
	stats["max_memory"] = redisInfo["maxmemory"]

	return s.evaluateRedisStats(redisInfo, stats)
}

// evaluateRedisStats evaluates the Redis server statistics and updates the stats map with relevant messages.
func (s *RedisComponent) evaluateRedisStats(redisInfo, stats map[string]string) map[string]string {
	poolSize := s.Client.Options().PoolSize
	poolStats := s.Client.PoolStats()

	stats["hits_connections"] = strconv.FormatUint(uint64(poolStats.Hits), 10)
	stats["misses_connections"] = strconv.FormatUint(uint64(poolStats.Misses), 10)
	stats["timeouts_connections"] = strconv.FormatUint(uint64(poolStats.Timeouts), 10)
	stats["total_connections"] = strconv.FormatUint(uint64(poolStats.TotalConns), 10)
	stats["idle_connections"] = strconv.FormatUint(uint64(poolStats.IdleConns), 10)
	stats["stale_connections"] = strconv.FormatUint(uint64(poolStats.StaleConns), 10)

	// Calculate the number of active connections.
	// Note: We use math.Max to ensure that activeConns is always non-negative,
	// avoiding the need for an explicit check for negative values.
	// This prevents a potential underflow situation.
	activeConns := uint64(math.Max(float64(poolStats.TotalConns-poolStats.IdleConns), 0))
	stats["active_connections"] = strconv.FormatUint(activeConns, 10)

	// Calculate the pool size percentage.
	connectedClients, _ := strconv.Atoi(redisInfo["connected_clients"])
	poolSizePercentage := float64(connectedClients) / float64(poolSize) * 100
	stats["pool_size_percentage"] = fmt.Sprintf("%.2f%%", poolSizePercentage)

	highConnectionThreshold := int(float64(poolSize) * 0.8)

	// Check if the number of connected clients is high.
	if connectedClients > highConnectionThreshold {
		stats["message"] = "Redis has a high number of connected clients"
	}

	// Check if the number of stale connections exceeds a threshold.
	minStaleConnectionsThreshold := 500
	if int(poolStats.StaleConns) > minStaleConnectionsThreshold {
		stats["message"] = fmt.Sprintf("Redis has %d stale connections.", poolStats.StaleConns)
	}

	// Check if Redis is using a significant amount of memory.
	usedMemory, _ := strconv.ParseInt(redisInfo["used_memory"], 10, 64)
	maxMemory, _ := strconv.ParseInt(redisInfo["maxmemory"], 10, 64)
	if maxMemory > 0 {
		usedMemoryPercentage := float64(usedMemory) / float64(maxMemory) * 100
		if usedMemoryPercentage >= 90 {
			stats["message"] = "Redis is using a significant amount of memory"
		}
	}

	// Check if Redis has been recently restarted.
	uptimeInSeconds, _ := strconv.ParseInt(redisInfo["uptime_in_seconds"], 10, 64)
	if uptimeInSeconds < 3600 {
		stats["message"] = "Redis has been recently restarted"
	}

	// Check if the number of idle connections is high.
	idleConns := int(poolStats.IdleConns)
	highIdleConnectionThreshold := int(float64(poolSize) * 0.7)
	if idleConns > highIdleConnectionThreshold {
		stats["message"] = "Redis has a high number of idle connections"
	}

	// Check if the connection pool utilization is high.
	poolUtilization := float64(poolStats.TotalConns-poolStats.IdleConns) / float64(poolSize) * 100
	highPoolUtilizationThreshold := 90.0
	if poolUtilization > highPoolUtilizationThreshold {
		stats["message"] = "Redis connection pool utilization is high"
	}

	return stats
}

// parseRedisInfo parses the Redis info response and returns a map of key-value pairs.
func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}
	return result
}
