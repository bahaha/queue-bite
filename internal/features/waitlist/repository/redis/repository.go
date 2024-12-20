package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	redis "github.com/redis/go-redis/v9"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/waitlist/domain"
)

var REDIS_WAITLIST = "waitlist/redis"

type redisWaitlistRepository struct {
	logger log.Logger
	client *redis.Client
	keys   *queueKeys
	ttl    time.Duration

	// preloaded Lua scripts
	joinScript     *redis.Script
	leaveScript    *redis.Script
	getPartyScript *redis.Script
}

func NewRedisWaitlistRepository(logger log.Logger, client *redis.Client, ttl time.Duration) *redisWaitlistRepository {
	return &redisWaitlistRepository{
		logger: logger,
		client: client,
		keys:   &queueKeys{},
		ttl:    ttl,

		joinScript:     redis.NewScript(joinScript),
		leaveScript:    redis.NewScript(leaveScript),
		getPartyScript: redis.NewScript(getPartyScript),
	}
}

func (r *redisWaitlistRepository) AddParty(ctx context.Context, party *domain.QueuedParty) (*domain.QueuedParty, error) {
	id := party.ID
	redisParty := newRedisQueuedParty(party)

	err := r.client.HSet(ctx, r.keys.partyDetails(id), redisParty).Err()
	if err != nil {
		r.logger.LogErr(REDIS_WAITLIST, err, "could not save the party detail", "party", party)
		return nil, err
	}

	joinKeys := []string{
		r.keys.waitingQueue(),
		r.keys.partyDetails(id),
		r.keys.waitTimePrefixsum(),
		r.keys.partyWaitTime(id),
	}
	joinArgs := []interface{}{
		id,
		party.EstimatedServiceTime.Round(time.Second),
		time.Now().Nanosecond(),
		r.ttl,
	}
	results, err := r.joinScript.Run(ctx, r.client, joinKeys, joinArgs...).Slice()
	if err != nil {
		r.logger.LogErr(REDIS_WAITLIST, err, "could not execute join waitlist script on redis", "keys", joinKeys, "args", joinArgs)
		return nil, fmt.Errorf("could not execute join waitlist script on redis: %w", err)
	}

	if results[0] == int64(0) {
		return nil, domain.ErrPartyAlreadyQueued
	}

	party.Position = int(results[1].(int64))
	party.EstimatedEndOfServiceTime = deserializeTime(results[2])

	return party, nil
}

func (r *redisWaitlistRepository) RemoveParty(ctx context.Context, partyID d.PartyID) error {
	leaveKeys := []string{
		r.keys.waitingQueue(),
		r.keys.partyDetails(partyID),
		r.keys.totalServiceTime(),
		r.keys.partyWaitTimePrefix(),
	}

	leaveArgs := []interface{}{partyID, "est"}

	results, err := r.leaveScript.Run(ctx, r.client, leaveKeys, leaveArgs...).Slice()
	if err != nil && err != redis.Nil {
		r.logger.LogDebug(REDIS_WAITLIST, "could not run leave queue script", "party id", partyID, "keys", leaveKeys, "args", leaveArgs)
		return fmt.Errorf("could not run leave queue script: %w", err)
	}

	if results == nil {
		err := domain.ErrPartyNotFound
		r.logger.LogErr(REDIS_WAITLIST, err, "inconsistency state: either could not find the party in the queue list or we could not find the estimated service time for the party", "party id", partyID)
		return err
	}

	position := results[0].(int64)
	estimatedServiceTimeOfLeftParty := results[1].(int64)

	r.logger.LogDebug(REDIS_WAITLIST, "party left the waitlist", "position", position, "estimated service time of party", estimatedServiceTimeOfLeftParty, "party id", partyID)
	return nil
}

func (r *redisWaitlistRepository) GetParty(ctx context.Context, partyID d.PartyID) (*domain.QueuedParty, error) {
	redisParty := &redisQueuedParty{}

	getPartyKeys := []string{
		r.keys.partyDetails(partyID),
		r.keys.waitingQueue(),
		r.keys.partyWaitTime(partyID),
		r.keys.totalServiceTime(),
	}
	getPartyArgs := []interface{}{partyID}
	results, err := r.getPartyScript.Run(ctx, r.client, getPartyKeys, getPartyArgs...).Slice()

	if err != nil && err != redis.Nil {
		r.logger.LogErr(REDIS_WAITLIST, err, "could not execute get queued entity scirpt", "party id", partyID)
		return nil, fmt.Errorf("could not execute get queued entity scirpt: %w", err)
	}

	if results == nil {
		r.logger.LogDebug(REDIS_WAITLIST, "could not found queued entity in consistent state", "party id", partyID, "err", err, "result", results)
		return nil, nil
	}

	cmd := redis.NewMapStringStringCmd(ctx)
	entityValue := sliceToMap(results[0].([]interface{}))
	cmd.SetVal(entityValue)

	if err := cmd.Scan(redisParty); err != nil {
		r.logger.LogDebug(REDIS_WAITLIST, "could not parse queued entity into redisParty struct", "party id", partyID, "value", entityValue)
		return nil, fmt.Errorf("could not parse queued entity into redisParty struct: %w", err)
	}

	estimatedServiceEndAt := deserializeTime(results[1])
	party := redisParty.asQueuedParty()
	party.EstimatedEndOfServiceTime = estimatedServiceEndAt
	party.Position = int(results[2].(int64))

	return party, nil
}

func (r *redisWaitlistRepository) GetQueueStatus(ctx context.Context) (*domain.QueueStatus, error) {
	waitNanos, err := r.client.Get(ctx, r.keys.waitTimePrefixsum()).Result()
	if err != nil && err != redis.Nil {
		r.logger.LogErr(REDIS_WAITLIST, err, "could not get the total wait from redis")
		return nil, fmt.Errorf("could not get the total wait from redis: %w", err)
	}
	totalWait := deserializeTime(waitNanos)

	amount, err := r.client.ZCard(ctx, r.keys.waitingQueue()).Result()
	if err != nil && err != redis.Nil {
		r.logger.LogErr(REDIS_WAITLIST, err, "could not find how many entities in the waitlist queue")
		return nil, fmt.Errorf("could not find how many entities in the waitlist queue: %w", err)
	}

	return &domain.QueueStatus{
		TotalParties:    int(amount),
		CurrentWaitTime: totalWait,
	}, nil
}

// deserializeTime converts Redis response (int64 or string in scientific notation)
// to time.Duration. It handles both numeric formats gracefully:
//   - "3e+11" -> 5m0s
//   - 300000000000 -> 5m0s
//
// Returns 0 duration if parsing fails.
func deserializeTime(val interface{}) time.Duration {
	switch v := val.(type) {
	case int64:
		return time.Duration(v)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return time.Duration(f)
	default:
		return 0
	}
}

func sliceToMap(slice []interface{}) map[string]string {
	m := make(map[string]string)
	for i := 0; i < len(slice); i += 2 {
		k := slice[i].(string)
		v := slice[i+1].(string)
		m[k] = v
	}
	return m
}
