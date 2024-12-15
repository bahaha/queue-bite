package server

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func WithRedisImage(envPrefix string) ServerComponent {
	return func(s *TestServer) error {
		ctx := context.Background()
		db, err := redis.Run(
			ctx,
			"docker.io/redis:7.2.4",
			redis.WithSnapshotting(10, 1),
			redis.WithLogLevel(redis.LogLevelNotice),
		)

		if err != nil {
			return err
		}

		host, err := db.Host(ctx)
		if err != nil {
			return err
		}
		port, err := db.MappedPort(ctx, "6379/tcp")
		if err != nil {
			return err
		}

		s.cleanups = append(s.cleanups, db.Terminate)
		s.envVars[fmt.Sprintf("%s_HOST", envPrefix)] = host
		s.envVars[fmt.Sprintf("%s_PORT", envPrefix)] = port.Port()

		return nil
	}
}
