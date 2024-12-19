package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"

	log "queue-bite/internal/config/logger"
	domain "queue-bite/internal/features/waitlist"
	"queue-bite/pkg/utils"
)

func TestJoinWaitlist(t *testing.T) {
	joinScriptHash := "ff40ec9ec852095a782a5ef8e4f29cac8cafdce6"

	tests := []struct {
		name       string
		party      *domain.Party
		queuePos   int64
		wantErr    bool
		validateFn func(t *testing.T, queued *QueuedParty)
	}{
		{
			name:     "validate lua script keys, args",
			party:    &domain.Party{Name: "Test Party", Size: 4},
			queuePos: 2,
			validateFn: func(t *testing.T, queued *QueuedParty) {
				assert.Equal(t, 5*time.Minute, queued.EstimatedServiceTime)
			},
		},
		{
			name:     "the second party into the queue",
			party:    &domain.Party{Name: "second party", Size: 2},
			queuePos: 1,
			validateFn: func(t *testing.T, queued *QueuedParty) {
				assert.Equal(t, 1, queued.QueuePosition)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, rmock := redismock.NewClientMock()
			service := NewRedisWaitlistService(
				log.NewNoopLogger(),
				client,
				newFixedServiceTimeEstimator(5),
			)
			uid := utils.GenerateUID()
			id := uid.String()
			service.(*redisWaitlistService).generateUID = func() ulid.ULID { return uid }

			rmock.ExpectEvalSha(joinScriptHash,
				[]string{"wq", fmt.Sprintf("qb:p:%s", id), "wq:twps", fmt.Sprintf("wq:ps:%s", id)},
				[]interface{}{id, 5 * time.Minute, uid.Time(), service.(*redisWaitlistService).waitTTL},
			).SetVal(tt.queuePos)
			rmock.ExpectHSet(fmt.Sprintf("qb:p:%s", id)).SetVal(1)

			queued, err := service.JoinQueue(context.Background(), tt.party)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, id, queued.ID)
			tt.validateFn(t, queued)
		})
	}
}

func TestGetQueuedParty(t *testing.T) {
	t.Parallel()
	var emptyQueuedParty func(mock redismock.ClientMock, partyID string)
	emptyQueuedParty = func(mock redismock.ClientMock, partyID string) {
		mock.ExpectHGetAll("qb:p:" + partyID).RedisNil()
	}

	var foundQueuedParty func(mock redismock.ClientMock, partyID string)
	foundQueuedParty = func(mock redismock.ClientMock, partyID string) {
		mock.ExpectHGetAll("qb:p:" + partyID).SetVal(map[string]string{
			"ID":                   "01JFDH85MYTFRWGJVV3PDSQR18",
			"Name":                 "CCC",
			"Size":                 "4",
			"JoinedAt":             "2024-12-19T02:42:27.102841+08:00",
			"EstimatedServiceTime": "300000000000",
		})
	}

	testCases := []struct {
		name     string
		setup    func(mock redismock.ClientMock, partyID string)
		validate func(t *testing.T, party *QueuedParty, err error)
	}{
		{
			name:  "party not found",
			setup: emptyQueuedParty,
			validate: func(t *testing.T, party *QueuedParty, err error) {
				assert.Error(t, err)
				assert.Nil(t, party)
			},
		},
		{
			name: "party details found but not in queue",
			setup: func(mock redismock.ClientMock, partyID string) {
				foundQueuedParty(mock, partyID)
				mock.ExpectZRank("wq", partyID).RedisNil()
			},
			validate: func(t *testing.T, party *QueuedParty, err error) {
				assert.NoError(t, err)
				assert.Nil(t, party)
			},
		},
		{
			name: "party found with position",
			setup: func(mock redismock.ClientMock, partyID string) {
				foundQueuedParty(mock, partyID)
				mock.ExpectZRank("wq", partyID).SetVal(2)
			},
			validate: func(t *testing.T, party *QueuedParty, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, party)
				assert.Equal(t, 2, party.QueuePosition)
				assert.Equal(t, "CCC", party.Name)
				assert.Equal(t, 4, party.Size)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, mock := redismock.NewClientMock()
			svc := &redisWaitlistService{
				client: client,
				logger: log.NewNoopLogger(),
			}

			partyID := "01JFDH85MYTFRWGJVV3PDSQR18"
			tc.setup(mock, partyID)
			party, err := svc.GetQueuedParty(context.Background(), partyID)
			tc.validate(t, party, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

type fixedServiceTimeEstimator struct {
	consumeTime int
}

func newFixedServiceTimeEstimator(minutes int) *fixedServiceTimeEstimator {
	return &fixedServiceTimeEstimator{
		consumeTime: minutes,
	}
}

func (e *fixedServiceTimeEstimator) EstimateServiceTime(ctx context.Context, payty *domain.Party) (time.Duration, error) {
	return time.Duration(e.consumeTime) * time.Minute, nil
}
