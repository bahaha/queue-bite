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
	logger    log.Logger
	client    *redis.Client
	keys      *queueKeys
	ttl       time.Duration
	scanRange int

	// preloaded Lua scripts
	joinScript     *redis.Script
	leaveScript    *redis.Script
	getPartyScript *redis.Script
}

func NewRedisWaitlistRepository(logger log.Logger, client *redis.Client, ttl time.Duration, scanRange int) *redisWaitlistRepository {
	return &redisWaitlistRepository{
		logger:    logger,
		client:    client,
		keys:      &queueKeys{},
		ttl:       ttl,
		scanRange: scanRange,

		joinScript:     redis.NewScript(joinScript),
		leaveScript:    redis.NewScript(leaveScript),
		getPartyScript: redis.NewScript(getPartyScript),
	}
}

func (r *redisWaitlistRepository) HasParty(ctx context.Context, partyID d.PartyID) bool {
	return r.client.Exists(ctx, r.keys.partyDetails(partyID)).Val() == int64(1)
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
		r.keys.totalServiceTime(),
		r.keys.waitingPartyCounter(),
	}
	joinArgs := []interface{}{
		id,
		int(party.EstimatedServiceTime.Seconds()),
		time.Now().Unix(),
		r.ttl,
		party.Status == d.PartyStatusWaiting,
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
		r.keys.waitTimePrefixsum(),
		r.keys.waitingPartyCounter(),
	}

	leaveArgs := []interface{}{partyID, "est", "status", d.PartyStatusWaiting}

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
	estimatedServiceTimeOfLeftParty := results[1]

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

func (r *redisWaitlistRepository) GetPartyDetails(ctx context.Context, partyID d.PartyID) (*domain.QueuedParty, error) {
	res := r.client.HGetAll(ctx, r.keys.partyDetails(partyID))
	if res.Err() != nil {
		r.logger.LogErr(REDIS_WAITLIST, res.Err(), "could not get party details", "party id", partyID)
		return nil, res.Err()
	}

	if res.Val() == nil {
		return nil, nil
	}

	queuedParty := &redisQueuedParty{}
	if err := res.Scan(queuedParty); err != nil {
		r.logger.LogErr(REDIS_WAITLIST, err, "could not parse queued entity into redisParty struct", "party id", partyID, "cmd", res)
		return nil, err
	}

	r.logger.LogDebug(REDIS_WAITLIST, "get party details", "party", queuedParty)
	return queuedParty.asQueuedParty(), nil
}

func (r *redisWaitlistRepository) GetQueueStatus(ctx context.Context) (*domain.QueueStatus, error) {
	waitSecs, err := r.client.Get(ctx, r.keys.waitTimePrefixsum()).Result()
	if err != nil && err != redis.Nil {
		r.logger.LogErr(REDIS_WAITLIST, err, "could not get the total wait from redis")
		return nil, fmt.Errorf("could not get the total wait from redis: %w", err)
	}
	totalWait := deserializeTime(waitSecs)

	totalServiceTime, err := r.client.Get(ctx, r.keys.totalServiceTime()).Result()
	if err != nil && err != redis.Nil {
		r.logger.LogErr(REDIS_WAITLIST, err, "could not get the total service time for those checked-in parties from redis")
		return nil, fmt.Errorf("could not get the total service time for those checked-in parties from redis: %w", err)
	}

	totalServiceSecs := deserializeTime(totalServiceTime)

	amount, err := r.client.ZCard(ctx, r.keys.waitingQueue()).Result()
	if err != nil && err != redis.Nil {
		r.logger.LogErr(REDIS_WAITLIST, err, "could not find how many entities in the waitlist queue")
		return nil, fmt.Errorf("could not find how many entities in the waitlist queue: %w", err)
	}

	waiting, err := r.client.Get(ctx, r.keys.waitingPartyCounter()).Int64()
	if err != nil && err != redis.Nil {
		r.logger.LogErr(REDIS_WAITLIST, err, "could not find how many party were waiting")
		return nil, fmt.Errorf("could not find how many party were waiting: %w", err)
	}

	return &domain.QueueStatus{
		TotalParties:    int(amount),
		WaitingParties:  int(waiting),
		CurrentWaitTime: totalWait - totalServiceSecs,
	}, nil
}

func (r *redisWaitlistRepository) ScanParties(ctx context.Context) (<-chan *domain.QueuedParty, error) {
	queuedParties := make(chan *domain.QueuedParty)

	// Start streaming in background
	go func() {
		defer close(queuedParties)

		offset := 0

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			ids, err := r.client.ZRange(ctx, r.keys.waitingQueue(), int64(offset), int64(offset+r.scanRange-1)).Result()
			if err != nil {
				r.logger.LogErr(REDIS_WAITLIST, err, "could not scan the waitlist", "from range", offset, "to", offset+r.scanRange-1)
				return
			}

			if len(ids) == 0 {
				return
			}

			for _, partyID := range ids {
				queuedParty, err := r.GetParty(ctx, d.PartyID(partyID))
				if err != nil {
					r.logger.LogErr(REDIS_WAITLIST, err, "error getting party", "party id", partyID)
					continue
				}

				select {
				case queuedParties <- queuedParty:
				case <-ctx.Done():
					return
				}
			}

			offset += len(ids)
		}
	}()

	return queuedParties, nil
}

func (r *redisWaitlistRepository) UpdatePartyStatus(ctx context.Context, partyID d.PartyID, status d.PartyStatus) error {
	originalStatus := r.client.HGet(ctx, r.keys.partyDetails(partyID), "status").Val()
	_, err := r.client.HSet(ctx, r.keys.partyDetails(partyID), "status", status).Result()
	if err != nil {
		r.logger.LogDebug(REDIS_WAITLIST, "could not found party in waitlist queue for status update", "party id", partyID)
	} else {
		r.logger.LogDebug(REDIS_WAITLIST, "update party status", "party id", partyID, "status", status)
	}

	if status == d.PartyStatusReady && originalStatus == string(d.PartyStatusWaiting) {
		r.client.IncrBy(ctx, r.keys.waitingPartyCounter(), -1)
	}
	return nil
}

// deserializeTime converts Redis response (int64 or string in seconds)
// to time.Duration. It handles both numeric formats gracefully:
// Returns 0 duration if parsing fails.
func deserializeTime(val interface{}) time.Duration {
	switch v := val.(type) {
	case int64:
		return time.Duration(v) * time.Second
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return time.Duration(f) * time.Second
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
