package service

import (
	"context"
	"fmt"
	"os"
	log "queue-bite/internal/config/logger"
	"queue-bite/internal/domain"
	hdr "queue-bite/internal/features/hostdesk/repository"
	hd "queue-bite/internal/features/hostdesk/service"
	st "queue-bite/internal/features/servicetime/service"
	wr "queue-bite/internal/features/waitlist/repository/redis"
	waitlist "queue-bite/internal/features/waitlist/service"
	"queue-bite/internal/platform/eventbus"
	ebr "queue-bite/internal/platform/eventbus/redis"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"
)

type testDeps struct {
	logger               log.Logger
	eventbus             eventbus.EventBus
	waitlist             waitlist.Waitlist
	hostdesk             hd.HostDesk
	maxOptimisticRetries int
}

func TestHandleNewPartyArrival(t *testing.T) {

	t.Run("Immediately Serving", func(t *testing.T) {
		ctx := context.Background()
		deps := setupTestDepdencies(t, 10)
		selection := NewOrderedSeatingStrategy(deps.waitlist)
		processing := NewInstantServingStrategy()
		service := NewSeatManager(deps.logger, deps.eventbus, deps.waitlist, deps.hostdesk, processing, selection, deps.maxOptimisticRetries)

		t.Run("serving success", func(t *testing.T) {
			queue, err := deps.waitlist.GetQueueStatus(ctx)
			require.NoError(t, err)
			assert.Equal(t, 0, queue.TotalParties)

			capacity, version, err := deps.hostdesk.GetCurrentCapacity(ctx)
			require.NoError(t, err)
			assert.Equal(t, 10, capacity)
			assert.Equal(t, 0, int(version))

			party := domain.NewParty("party-1", "name", 8)
			result, err := service.ProcessNewParty(ctx, party)
			assert.Equal(t, domain.PartyStatusServing, result.Status)
		})

		t.Run("serving if capacity still enough", func(t *testing.T) {
			queue, err := deps.waitlist.GetQueueStatus(ctx)
			require.NoError(t, err)
			assert.Equal(t, 0, queue.TotalParties)

			capacity, _, err := deps.hostdesk.GetCurrentCapacity(ctx)
			require.NoError(t, err)
			assert.Equal(t, 2, capacity)

			party := domain.NewParty("party-2", "name", 1)
			result, err := service.ProcessNewParty(ctx, party)
			assert.Equal(t, domain.PartyStatusServing, result.Status)
		})

		t.Run("wait in queue if capacity is insufficient", func(t *testing.T) {
			queue, err := deps.waitlist.GetQueueStatus(ctx)
			require.NoError(t, err)
			assert.Equal(t, 0, queue.TotalParties)

			capacity, _, err := deps.hostdesk.GetCurrentCapacity(ctx)
			require.NoError(t, err)
			assert.Equal(t, 1, capacity)

			party := domain.NewParty("party-3", "name", 4)
			result, err := service.ProcessNewParty(ctx, party)
			assert.Equal(t, domain.PartyStatusWaiting, result.Status)
		})

		t.Run("wait in queue even if the capacity is enough but there waiting parties in queue", func(t *testing.T) {
			queue, err := deps.waitlist.GetQueueStatus(ctx)
			require.NoError(t, err)
			assert.Equal(t, 1, queue.TotalParties)

			capacity, _, err := deps.hostdesk.GetCurrentCapacity(ctx)
			require.NoError(t, err)
			assert.Equal(t, 1, capacity)

			party := domain.NewParty("party-4", "name", 1)
			result, err := service.ProcessNewParty(ctx, party)
			assert.Equal(t, domain.PartyStatusWaiting, result.Status)
		})
	})

	t.Run("Queued First", func(t *testing.T) {
		ctx := context.Background()
		deps := setupTestDepdencies(t, 10)
		selection := NewOrderedSeatingStrategy(deps.waitlist)
		processing := NewQueueFirstStrategy()
		service := NewSeatManager(deps.logger, deps.eventbus, deps.waitlist, deps.hostdesk, processing, selection, deps.maxOptimisticRetries)

		t.Run("ready to check in", func(t *testing.T) {
			queue, err := deps.waitlist.GetQueueStatus(ctx)
			require.NoError(t, err)
			assert.Equal(t, 0, queue.TotalParties)

			capacity, version, err := deps.hostdesk.GetCurrentCapacity(ctx)
			require.NoError(t, err)
			assert.Equal(t, 10, capacity)
			assert.Equal(t, 0, int(version))

			party := domain.NewParty("party-1", "name", 8)
			result, err := service.ProcessNewParty(ctx, party)
			assert.Equal(t, domain.PartyStatusReady, result.Status)

			queue, err = deps.waitlist.GetQueueStatus(ctx)
			require.NoError(t, err)
			assert.Equal(t, 1, queue.TotalParties)
		})

		t.Run("ready to check in in the queue if capacity still enough", func(t *testing.T) {
			queue, err := deps.waitlist.GetQueueStatus(ctx)
			require.NoError(t, err)
			assert.Equal(t, 1, queue.TotalParties)

			capacity, _, err := deps.hostdesk.GetCurrentCapacity(ctx)
			require.NoError(t, err)
			assert.Equal(t, 2, capacity)

			party := domain.NewParty("party-2", "name", 1)
			result, err := service.ProcessNewParty(ctx, party)
			assert.Equal(t, domain.PartyStatusReady, result.Status)
		})

		t.Run("wait in queue if capacity is insufficient", func(t *testing.T) {
			queue, err := deps.waitlist.GetQueueStatus(ctx)
			require.NoError(t, err)
			assert.Equal(t, 2, queue.TotalParties)

			capacity, _, err := deps.hostdesk.GetCurrentCapacity(ctx)
			require.NoError(t, err)
			assert.Equal(t, 1, capacity)

			party := domain.NewParty("party-3", "name", 4)
			result, err := service.ProcessNewParty(ctx, party)
			assert.Equal(t, domain.PartyStatusWaiting, result.Status)
		})

		// FIXME: queue first strategy needs to know more queue details like how many parties is waiting, current context have only how many parties in the queue
		// t.Run("wait in queue even if the capacity is enough but there waiting parties in queue", func(t *testing.T) {
		// 	queue, err := deps.waitlist.GetQueueStatus(ctx)
		// 	require.NoError(t, err)
		// 	assert.Equal(t, 3, queue.TotalParties)
		//
		// 	capacity, _, err := deps.hostdesk.GetCurrentCapacity(ctx)
		// 	require.NoError(t, err)
		// 	assert.Equal(t, 1, capacity)
		//
		// 	party := domain.NewParty("party-4", "name", 1)
		// 	result, err := service.ProcessNewParty(ctx, party)
		// 	assert.Equal(t, domain.PartyStatusWaiting, result.Status)
		// })
	})

	t.Run("check in failed will fallback to ready in the waitlist", func(t *testing.T) {
		ctx := context.Background()
		hostdesk := hd.NewMockHostDesk(gomock.NewController(t))
		deps := setupTestDepdencies(t, 10)
		selection := NewOrderedSeatingStrategy(deps.waitlist)
		processing := NewInstantServingStrategy()
		service := NewSeatManager(deps.logger, deps.eventbus, deps.waitlist, hostdesk, processing, selection, deps.maxOptimisticRetries)

		hostdesk.
			EXPECT().
			GetCurrentCapacity(ctx).
			Return(10, domain.Version(0), nil).
			AnyTimes()

		hostdesk.
			EXPECT().
			PreserveSeats(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
			Return(true, nil).
			AnyTimes()

		hostdesk.
			EXPECT().
			CheckIn(ctx, gomock.Any()).
			Return(fmt.Errorf("unexpected runtime error")).
			AnyTimes()

		party := domain.NewParty("party-1", "name", 8)
		result, err := service.ProcessNewParty(ctx, party)
		require.NoError(t, err)
		assert.Equal(t, domain.PartyStatusReady, result.Status)
	})

	t.Run("concurrent optimistic lock", func(t *testing.T) {

		t.Run("retry failed", func(t *testing.T) {
			ctx := context.Background()
			hostdesk := hd.NewMockHostDesk(gomock.NewController(t))
			deps := setupTestDepdencies(t, 10)
			selection := NewOrderedSeatingStrategy(deps.waitlist)
			processing := NewQueueFirstStrategy()
			service := NewSeatManager(deps.logger, deps.eventbus, deps.waitlist, hostdesk, processing, selection, deps.maxOptimisticRetries)
			hostdesk.
				EXPECT().
				GetCurrentCapacity(ctx).
				Return(10, domain.Version(0), nil).
				AnyTimes()

			hostdesk.
				EXPECT().
				PreserveSeats(ctx, gomock.Eq(domain.PartyID("party-1")), gomock.Eq(8), gomock.Eq(domain.Version(0))).
				Return(false, domain.ErrVersionMismatch).
				AnyTimes()

			party := domain.NewParty("party-1", "name", 8)
			_, err := service.ProcessNewParty(ctx, party)
			assert.ErrorIs(t, err, domain.ErrTooManyOptimisticLockRetries)
		})

		t.Run("retry success", func(t *testing.T) {
			ctx := context.Background()
			hostdesk := hd.NewMockHostDesk(gomock.NewController(t))
			deps := setupTestDepdencies(t, 10)
			selection := NewOrderedSeatingStrategy(deps.waitlist)
			processing := NewQueueFirstStrategy()
			service := NewSeatManager(deps.logger, deps.eventbus, deps.waitlist, hostdesk, processing, selection, deps.maxOptimisticRetries)
			gomock.InOrder(
				hostdesk.
					EXPECT().
					GetCurrentCapacity(ctx).
					Return(10, domain.Version(0), nil).
					Times(1),
				hostdesk.
					EXPECT().
					GetCurrentCapacity(ctx).
					Return(8, domain.Version(1), nil).
					AnyTimes(),
			)

			gomock.InOrder(
				hostdesk.
					EXPECT().
					PreserveSeats(ctx, gomock.Eq(domain.PartyID("party-1")), gomock.Eq(8), gomock.Eq(domain.Version(0))).
					Return(false, domain.ErrVersionMismatch).
					Times(1),
				hostdesk.
					EXPECT().
					PreserveSeats(ctx, gomock.Eq(domain.PartyID("party-1")), gomock.Eq(8), gomock.Eq(domain.Version(1))).
					Return(true, nil).
					Times(1),
			)

			party := domain.NewParty("party-1", "name", 8)
			result, err := service.ProcessNewParty(ctx, party)
			require.NoError(t, err)
			assert.Equal(t, domain.PartyStatusReady, result.Status)
		})
	})

}

func setupTestDepdencies(t *testing.T, seats int) *testDeps {
	redisClient, cleanup := setupRedisContainer(t)
	t.Cleanup(cleanup)
	logger := log.NewZerologLogger(os.Stdout, true)
	registry := eventbus.NewEventRegistry()
	eventbus := ebr.NewRedisEventBus(logger, redisClient, registry)
	waitlist := waitlist.NewWaitlistService(
		logger,
		wr.NewRedisWaitlistRepository(logger, redisClient, 5*time.Second, 5),
		st.NewFixedRateEstimator(1*time.Minute),
		eventbus,
	)
	hostdesk := hd.NewInstantServeHostDesk(logger, seats, hdr.NewInMemoryHostDeskRepository(logger), eventbus, nil)
	maxOptimisticRetries := 3

	return &testDeps{
		logger:               logger,
		eventbus:             eventbus,
		waitlist:             waitlist,
		hostdesk:             hostdesk,
		maxOptimisticRetries: maxOptimisticRetries,
	}
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
