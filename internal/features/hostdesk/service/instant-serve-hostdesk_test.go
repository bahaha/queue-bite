package service

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/hostdesk/domain"
	"queue-bite/internal/features/hostdesk/repository"
	"queue-bite/internal/platform/eventbus"
	ebr "queue-bite/internal/platform/eventbus/redis"
)

func TestPreserveSeats(t *testing.T) {
	logger := log.NewZerologLogger(os.Stdout, true)
	redisClient, cleanup := setupRedisContainer(t)
	t.Cleanup(cleanup)

	inmemoryRepo := repository.NewInMemoryHostDeskRepository(logger)
	redisRepo := repository.NewRedisHostDeskRepository(logger, redisClient)
	registry := eventbus.NewEventRegistry()
	eventbus := ebr.NewRedisEventBus(logger, redisClient, registry)
	totalSeats := 12
	impl := []repository.HostDeskRepository{inmemoryRepo, redisRepo}
	svc := []HostDesk{}
	for _, repo := range impl {
		svc = append(svc, NewInstantServeHostDesk(logger, totalSeats, repo, eventbus, nil))
	}

	t.Run("successful reservation", func(t *testing.T) {
		for _, service := range svc {
			ok, err := service.PreserveSeats(context.Background(), "party-1", 10, 0)
			require.NoError(t, err)
			assert.True(t, ok)

			available, version, err := service.GetCurrentCapacity(context.Background())
			require.NoError(t, err)
			assert.Equal(t, 2, available)
			assert.Equal(t, 1, int(version))
		}
	})

	t.Run("party already exists", func(t *testing.T) {
		for _, service := range svc {
			ok, err := service.PreserveSeats(context.Background(), "party-1", 2, 1)
			assert.ErrorIs(t, err, domain.ErrPartyAlreadyExists)
			assert.False(t, ok)
		}
	})

	t.Run("insufficient seats", func(t *testing.T) {
		for _, service := range svc {
			ok, err := service.PreserveSeats(context.Background(), "party-2", 4, 1)
			assert.ErrorIs(t, err, domain.ErrInsufficientCapacity)
			assert.False(t, ok)
		}
	})

	t.Run("version dismatch", func(t *testing.T) {
		for _, service := range svc {
			ok, err := service.PreserveSeats(context.Background(), "party-2", 2, 0)
			assert.ErrorIs(t, err, d.ErrVersionMismatch)
			assert.False(t, ok)
		}
	})
}

func setupRedisContainer(t *testing.T) (*redis.Client, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)

	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(t, err)

	cleanup := func() {
		require.NoError(t, container.Terminate(ctx))
	}

	client := redis.NewClient(&redis.Options{Addr: endpoint})
	return client, cleanup
}
