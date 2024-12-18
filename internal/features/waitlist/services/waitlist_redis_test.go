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
	client, rmock := redismock.NewClientMock()
	joinWaitlistScriptSha := "a8a0f14b5b1c53f7778a38c2f359355dee131e94"

	var mockJoinScriptReturns func(uid ulid.ULID, waitTTL time.Duration, val int64)
	mockJoinScriptReturns = func(uid ulid.ULID, waitTTL time.Duration, val int64) {
		id := uid.String()
		rmock.ExpectEvalSha(joinWaitlistScriptSha,
			[]string{
				"wq",
				fmt.Sprintf("qb:p:%s", id),
				"wq:twps",
				fmt.Sprintf("wq:ps:%s", id),
			},
			[]interface{}{
				id,
				5 * time.Minute,
				uid.Time(),
				waitTTL,
			},
		).SetVal(val)
	}

	t.Run("validate lua script keys, args", func(t *testing.T) {
		t.Parallel()
		service := NewRedisWaitlistService(
			log.NewNoopLogger(),
			client,
			newFixedServiceTimeEstimator(5),
		)
		uid := utils.GenerateUID()
		id := uid.String()
		service.(*redisWaitlistService).generateUID = func() ulid.ULID {
			return uid
		}

		party := &domain.Party{Name: "Test Party", Size: 4}
		mockJoinScriptReturns(uid, service.(*redisWaitlistService).waitTTL, 2)
		rmock.ExpectHSet(fmt.Sprintf("qb:p:%s", id)).SetVal(1)

		queued, err := service.JoinQueue(context.Background(), party)
		assert.NoError(t, err)
		assert.Equal(t, id, queued.ID)
		assert.Equal(t, 5*time.Minute, queued.EstimatedServiceTime)
	})

	t.Run("the second party into the queue", func(t *testing.T) {
		t.Parallel()
		uid := utils.GenerateUID()
		service := NewRedisWaitlistService(
			log.NewNoopLogger(),
			client,
			newFixedServiceTimeEstimator(5),
		)
		ttl := service.(*redisWaitlistService).waitTTL
		service.(*redisWaitlistService).generateUID = func() ulid.ULID {
			return uid
		}

		party := &domain.Party{Name: "second party", Size: 2}
		mockJoinScriptReturns(uid, ttl, 1)

		queued, err := service.JoinQueue(context.Background(), party)
		assert.NoError(t, err)
		assert.Equal(t, 1, queued.QueuePosition)
	})
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
