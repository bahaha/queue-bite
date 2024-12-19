package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	log "queue-bite/internal/config/logger"
	domain "queue-bite/internal/features/waitlist"
	"queue-bite/pkg/utils"
)

func TestJoinWaitlist(t *testing.T) {
	joinScriptHash := "84cde3f5aaa51cb273001e2791dce3b5a1581c93"

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
				assert.Equal(t, 10*time.Minute, queued.EstimatedWait)
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

			totalWait := time.Duration(tt.queuePos*5) * time.Minute
			rmock.ExpectEvalSha(joinScriptHash,
				[]string{"wq", fmt.Sprintf("qb:p:%s", id), "wq:twps", fmt.Sprintf("wq:ps:%s", id)},
				[]interface{}{id, 5 * time.Minute, uid.Time(), service.(*redisWaitlistService).waitTTL},
			).SetVal([]interface{}{tt.queuePos, fmt.Sprintf("%.0e", float64(totalWait.Nanoseconds()))})
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

	const partyID = "01JFDH85MYTFRWGJVV3PDSQR18"
	const scriptSha = "bb0ebe03f23b23bc3b8eafa9aecc7f7fbc53c573"

	testCases := []struct {
		name      string
		scriptRet interface{}
		scriptErr error
		validate  func(t *testing.T, party *QueuedParty, err error)
	}{
		{
			name:      "party not found",
			scriptErr: redis.Nil,
			validate: func(t *testing.T, party *QueuedParty, err error) {
				assert.NoError(t, err)
				assert.Nil(t, party)
			},
		},
		{
			name:      "script execution error",
			scriptErr: errors.New("redis connection error"),
			validate: func(t *testing.T, party *QueuedParty, err error) {
				assert.Error(t, err)
				assert.Nil(t, party)
			},
		},
		{
			name: "successful party lookup",
			scriptRet: []interface{}{
				[]interface{}{
					"ID", "01JFDH85MYTFRWGJVV3PDSQR18",
					"Name", "CCC",
					"Size", "4",
					"JoinedAt", "2024-12-19T02:42:27.102841+08:00",
					"EstimatedServiceTime", "300000000000",
				},
				fmt.Sprintf("%.0e", float64(10*time.Minute.Nanoseconds())), // EstimatedWait
				int64(2), // QueuePosition
			},
			validate: func(t *testing.T, party *QueuedParty, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, party)
				assert.Equal(t, 2, party.QueuePosition, "party queue qosition")
				assert.Equal(t, "CCC", party.Name)
				assert.Equal(t, 4, party.Size)
				assert.Equal(t, 10*time.Minute, party.EstimatedWait)
				assert.Equal(t, 5*time.Minute, party.EstimatedServiceTime)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, mock := redismock.NewClientMock()

			expectedCmd := mock.ExpectEvalSha(scriptSha,
				[]string{"qb:p:" + partyID, "wq", "wq:ps:" + partyID},
				[]interface{}{partyID},
			)
			if tc.scriptErr != nil {
				expectedCmd.SetErr(tc.scriptErr)
			} else {
				expectedCmd.SetVal(tc.scriptRet)
			}

			svc := NewRedisWaitlistService(
				log.NewNoopLogger(),
				client,
				newFixedServiceTimeEstimator(5),
			)

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
