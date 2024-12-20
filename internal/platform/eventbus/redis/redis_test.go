package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	log "queue-bite/internal/config/logger"
	"queue-bite/internal/platform/eventbus"
)

type TestEvent struct {
	ID      string    `json:"id"`
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

func (e *TestEvent) Topic() string {
	return "test.event"
}

func (e *TestEvent) NewEvent() eventbus.Event {
	return &TestEvent{}
}

func TestRedisEventBusSerializeDeserialize(t *testing.T) {
	endpoint, cleanup := setupRedisContainer(t)
	defer cleanup()

	client := redis.NewClient(&redis.Options{Addr: endpoint})
	defer client.Close()

	registry := eventbus.NewEventRegistry()
	registry.Register("test.event", &TestEvent{})

	bus := NewRedisEventBus(log.NewNoopLogger(), client, registry)
	ctx := context.Background()

	originalEvent := &TestEvent{ID: "test-123", Time: time.Now().UTC(), Message: "test message"}

	receivedEvents := make(chan eventbus.Event, 1)
	bus.Subscribe("test.event", func(ctx context.Context, event eventbus.Event) error {
		receivedEvents <- event
		return nil
	})

	err := bus.Publish(ctx, originalEvent)
	assert.NoError(t, err)

	select {
	case receivedEvent := <-receivedEvents:
		received, ok := receivedEvent.(*TestEvent)
		assert.True(t, ok)

		assert.Equal(t, originalEvent.ID, received.ID)
		assert.Equal(t, originalEvent.Message, received.Message)
		assert.Equal(t, originalEvent.Time.Unix(), received.Time.Unix())

	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for event")
	}

	t.Cleanup(func() {
		err := bus.sub.Close()
		require.NoError(t, err)
	})
}

func setupRedisContainer(t *testing.T) (string, func()) {
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

	return endpoint, cleanup
}
