package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jinzhu/copier"
	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"

	log "queue-bite/internal/config/logger"
	domain "queue-bite/internal/features/waitlist"
	"queue-bite/pkg/utils"
)

const (
	WAIT_REDIS_IMPL = "waitlist/redis_impl"
)

type redisWaitlistService struct {
	logger               log.Logger
	client               *redis.Client
	serviceTimeEstimator ServiceTimeEstimator
	waitTTL              time.Duration

	// atomic Lua script operations
	joinScript *redis.Script

	generateUID func() ulid.ULID
}

func NewRedisWaitlistService(logger log.Logger, client *redis.Client, serviceTimeEstimator ServiceTimeEstimator) WaitlistService {
	return &redisWaitlistService{
		logger:               logger,
		client:               client,
		serviceTimeEstimator: serviceTimeEstimator,
		waitTTL:              24 * time.Hour,

		joinScript:  redis.NewScript(joinWaitlistScript),
		generateUID: utils.GenerateUID,
	}
}

type QueuedEntity struct {
	ID                   string        `redis:"ID"`
	Name                 string        `redis:"Name"`
	Size                 int           `redis:"Size"`
	QueuePosition        int           `redis:"-"`
	JoinedAt             time.Time     `redis:"JoinedAt"`
	EstimatedServiceTime time.Duration `redis:"EstimatedServiceTime"`
}

const joinWaitlistScript = `
local waitlist_key = KEYS[1]
local party_detail_key = KEYS[2]
local total_wait_prefixsum_key = KEYS[3]
local party_wait_prefixsum_key = KEYS[4]
local party_id = ARGV[1]
local estimated_service_time = ARGV[2]
local join_score = ARGV[3]
local ttl = ARGV[4]

-- Add to sorted set
local wait_entries_ahead = redis.call('ZCARD', waitlist_key)
redis.call('ZADD', waitlist_key, join_score, party_id)
redis.call('EXPIRE', waitlist_key, ttl)


-- Increment total wait time and get new value
local next_wait = redis.call('INCRBY', total_wait_prefixsum_key, estimated_service_time)

-- Set wait time prefix sum for this group with TTL
redis.call('SET', party_wait_prefixsum_key, next_wait, 'EX', ttl)

return wait_entries_ahead
`

// JoinQueue adds a party to the waitlist queue and returns with their queue position and estimated wait time.
// It stores the party details in Redis using several data structures:
//   - Sorted set for maintaining queue order (wq)
//   - Hash for party details (qb:p:<id>)
//   - Total prefixsum for total wait time (wq:twqs), which is helpful for the visitors were not queued.
//   - prefixsum for a specific party (wq:ps:<id>), which is helpful for the party to aware the remaining wait time.
//
// Example Redis state after joining:
//
//		ZSET waitlist
//		    "ulid1" -> 1639980000
//
//		HASH qb:p
//		    ID -> "ulid1"
//		    Name -> "CC"
//		    Size -> "2"
//		    JoinedAt -> "2023-12-19T15:00:00Z"
//		    EstimatedServiceTime -> "1800000000000"    // The total time remaining to wait for this party
//
//	 The Lua script will return the parties ahead of this new joining party, which is the order in the queue.
func (s *redisWaitlistService) JoinQueue(ctx context.Context, party *domain.Party) (*QueuedParty, error) {
	queued := &QueuedEntity{}
	copier.Copy(queued, party)

	uid := s.generateUID()
	id := uid.String()
	queued.ID = id
	queued.JoinedAt = time.Now()
	// TODO: timeout for estimate wait time for a party
	waitTime, err := s.serviceTimeEstimator.EstimateServiceTime(ctx, party)
	if err != nil {
		s.logger.LogErr(WAIT_ESTIMATOR, err, "could not estimate wait time for party", "name", party.Name, "size", party.Size)
		return nil, err
	}
	queued.EstimatedServiceTime = waitTime
	s.client.HSet(ctx, s.getPartyDetailKey(queued.ID), queued)

	joinKeys := []string{
		s.getWaitlistKey(),               // local waitlist_key = KEYS[1]
		s.getPartyDetailKey(id),          // local party_detail_key = KEYS[2]
		s.getTotalWaitTimePrefixSumKey(), // local total_wait_prefixsum_key = KEYS[3]
		s.getWaitTimePrefixSumKey(id),    // local party_wait_prefixsum_key = KEYS[4]
	}

	joinArgs := []interface{}{
		id,                          // local party_id = ARGV[1]
		waitTime.Round(time.Second), // local estimated_service_time = ARGV[2]
		uid.Time(),                  // local join_score = ARGV[3]
		s.waitTTL,                   // local ttl = ARGV[4]
	}

	val, err := s.joinScript.Run(ctx, s.client, joinKeys, joinArgs...).Result()
	if err != nil {
		s.logger.LogErr(WAIT_REDIS_IMPL, err, "could not execute join waitlist script on redis", "keys", joinKeys, "args", joinArgs)
		return nil, err
	}

	if partiesAhead, ok := val.(int64); ok {
		queued.QueuePosition = int(partiesAhead)
	}

	s.logger.LogDebug(WAIT_REDIS_IMPL, "join waitlist success", "party detail", queued)
	queuedParty := &QueuedParty{}
	copier.Copy(queuedParty, queued)
	return queuedParty, nil
}

func (s *redisWaitlistService) getWaitlistKey() string {
	return "wq"
}

func (s *redisWaitlistService) getPartyDetailKey(partyID string) string {
	return fmt.Sprintf("qb:p:%s", partyID)
}

func (s *redisWaitlistService) getWaitTimePrefixSumKey(partyID string) string {
	return fmt.Sprintf("wq:ps:%s", partyID)
}

func (s *redisWaitlistService) getTotalWaitTimePrefixSumKey() string {
	return "wq:twps"
}
