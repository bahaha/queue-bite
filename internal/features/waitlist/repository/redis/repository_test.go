package redis

import (
	"context"
	"os"
	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/waitlist/domain"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRedisWaitlistRepository(t *testing.T) {
	endpoint, cleanup := setupRedisContainer(t)
	defer cleanup()

	client := redis.NewClient(&redis.Options{Addr: endpoint})
	defer client.Close()

	repo := NewRedisWaitlistRepository(log.NewZerologLogger(os.Stdout, true), client, 1*time.Minute, 2)
	ctx := context.Background()

	party := &domain.QueuedParty{
		Party: &d.Party{
			ID:                   "test-party-1",
			Name:                 "test-party-name",
			Status:               d.PartyStatusWaiting,
			Size:                 4,
			EstimatedServiceTime: 30 * time.Minute,
		},
	}
	partyII := &domain.QueuedParty{}
	copier.Copy(partyII, party)
	partyII.ID = "test-party-2"
	partyII.EstimatedServiceTime = 15 * time.Minute
	partyIII := &domain.QueuedParty{}
	copier.Copy(partyIII, party)
	partyIII.ID = "test-party-3"
	partyIII.EstimatedServiceTime = 5 * time.Minute

	t.Run("add and retrieve party", func(t *testing.T) {
		addedParty, err := repo.AddParty(ctx, party)
		require.NoError(t, err)
		assert.Equal(t, 0, addedParty.Position)
		assert.Equal(t, party.EstimatedServiceTime, addedParty.EstimatedEndOfServiceTime)
		assert.Equal(t, d.PartyStatusWaiting, addedParty.Status)
		assert.Equal(t, time.Duration(0), addedParty.RemainingWaitTime())

		retrievedParty, err := repo.GetParty(ctx, party.ID)
		require.NoError(t, err)
		assert.Equal(t, party.ID, retrievedParty.ID)
		assert.Equal(t, party.Name, retrievedParty.Name)
		assert.Equal(t, party.Size, retrievedParty.Size)
		assert.Equal(t, party.EstimatedServiceTime, retrievedParty.EstimatedServiceTime)
		assert.Equal(t, d.PartyStatusWaiting, retrievedParty.Status)
		assert.Equal(t, 0, retrievedParty.Position)
		assert.Equal(t, time.Duration(0), addedParty.RemainingWaitTime())
	})

	t.Run("new party join the waitlist", func(t *testing.T) {
		addedPartyII, err := repo.AddParty(ctx, partyII)
		require.NoError(t, err)
		assert.Equal(t, 1, addedPartyII.Position)
		assert.Equal(t, party.EstimatedServiceTime+partyII.EstimatedServiceTime, addedPartyII.EstimatedEndOfServiceTime)
		assert.Equal(t, party.EstimatedServiceTime, addedPartyII.RemainingWaitTime())

		retrievedPartyII, err := repo.GetParty(ctx, partyII.ID)
		require.NoError(t, err)
		assert.Equal(t, addedPartyII.Position, retrievedPartyII.Position)
		assert.Equal(t, party.EstimatedServiceTime+partyII.EstimatedServiceTime, retrievedPartyII.EstimatedEndOfServiceTime)

		addedPartyIII, err := repo.AddParty(ctx, partyIII)
		require.NoError(t, err)
		assert.Equal(t, 2, addedPartyIII.Position)
		assert.Equal(t, party.EstimatedServiceTime+partyII.EstimatedServiceTime+partyIII.EstimatedServiceTime, addedPartyIII.EstimatedEndOfServiceTime)
		assert.Equal(t, party.EstimatedServiceTime+partyII.EstimatedServiceTime, addedPartyIII.RemainingWaitTime())

		retrievedPartyIII, err := repo.GetParty(ctx, partyIII.ID)
		require.NoError(t, err)
		assert.Equal(t, addedPartyIII.Position, retrievedPartyIII.Position)
	})

	t.Run("a queued party join again", func(t *testing.T) {
		queuedParty, err := repo.AddParty(ctx, party)
		assert.ErrorIs(t, err, domain.ErrPartyAlreadyQueued)
		assert.Nil(t, queuedParty)
	})

	t.Run("queue status reflects total parties and wait time", func(t *testing.T) {
		status, err := repo.GetQueueStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, 3, status.TotalParties)
		assert.Equal(t, party.EstimatedServiceTime+partyII.EstimatedServiceTime+partyIII.EstimatedServiceTime, status.CurrentWaitTime)
	})

	t.Run("scan queued parties by order in queue", func(t *testing.T) {
		parties, err := repo.ScanParties(ctx)
		require.NoError(t, err)
		idx := 0

		expectedIDs := []string{"test-party-1", "test-party-2", "test-party-3"}
		for party := range parties {
			assert.Equal(t, d.PartyID(expectedIDs[idx]), party.ID)
			idx++
		}
	})

	t.Run("party in the middle removed", func(t *testing.T) {
		retrievedPartyIII, err := repo.GetParty(ctx, partyIII.ID)

		err = repo.RemoveParty(ctx, partyII.ID)
		require.NoError(t, err)

		retrievedPartyIII, err = repo.GetParty(ctx, partyIII.ID)
		assert.Equal(t, 1, retrievedPartyIII.Position)
		assert.Equal(t, party.EstimatedServiceTime, retrievedPartyIII.RemainingWaitTime())
	})

	t.Run("notify the head of the waitlist for ready to serve", func(t *testing.T) {
		err := repo.UpdatePartyStatus(ctx, party.ID, d.PartyStatusReady)
		require.NoError(t, err)

		retrievedParty, err := repo.GetPartyDetails(ctx, party.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrievedParty)
		assert.Equal(t, d.PartyStatusReady, retrievedParty.Status)
	})

	t.Run("remove the first party in queue", func(t *testing.T) {
		err := repo.RemoveParty(ctx, party.ID)
		require.NoError(t, err)

		retrievedPartyIII, err := repo.GetParty(ctx, partyIII.ID)
		assert.Equal(t, 0, retrievedPartyIII.Position)
		assert.Equal(t, time.Duration(0), retrievedPartyIII.RemainingWaitTime())
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
