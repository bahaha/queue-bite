package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jinzhu/copier"
	"github.com/redis/go-redis/v9"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/hostdesk/domain"
)

var REDIS_HOSTDESK = "hostdesk/redis"
var SKIP_VERSION_CHECK = -1

type hostdeskRedisKeys struct{}

// hd:state:<party_id>
func (k *hostdeskRedisKeys) getPartyStateKey(partyID d.PartyID) string {
	return fmt.Sprintf("hd:state:%s", partyID)
}

// hd:stats
func (k *hostdeskRedisKeys) getStatsKey() string {
	return "hd:stats"
}

type hostdeskStatsHash struct {
	Occupied  int `redis:"Occupied"`
	Preserved int `redis:"Preserved"`
	Version   int `redis:"Version"`
}

type RedisHostDeskRepository struct {
	logger log.Logger
	client *redis.Client
	keys   *hostdeskRedisKeys
}

func NewRedisHostDeskRepository(logger log.Logger, client *redis.Client) HostDeskRepository {
	repo := &RedisHostDeskRepository{
		logger: logger,
		client: client,
		keys:   &hostdeskRedisKeys{},
	}

	ctx := context.Background()
	exists, _ := client.HExists(ctx, repo.keys.getStatsKey(), "Version").Result()
	if !exists {
		client.HMSet(ctx, repo.keys.getStatsKey(), &hostdeskStatsHash{Occupied: 0, Preserved: 0, Version: 0})
	}

	return repo
}

func (r *RedisHostDeskRepository) GetOccupiedSeats(ctx context.Context) (int, error) {
	occupied, err := r.client.HGet(ctx, r.keys.getStatsKey(), "Occupied").Int()
	if err == redis.Nil {
		return 0, nil
	}
	return occupied, err
}

func (r *RedisHostDeskRepository) GetPreservedSeats(ctx context.Context) (int, error) {
	preserved, err := r.client.HGet(ctx, r.keys.getStatsKey(), "Preserved").Int()
	if err == redis.Nil {
		return 0, nil
	}
	return preserved, err
}

func (r *RedisHostDeskRepository) GetTotalSeatsInUse(ctx context.Context) (int, d.Version, error) {
	res := r.client.HGetAll(ctx, r.keys.getStatsKey())
	if res.Err() != nil {
		return 0, 0, res.Err()
	}
	stats := &hostdeskStatsHash{}
	if err := res.Scan(stats); err != nil {
		return 0, 0, err
	}

	return stats.Occupied + stats.Preserved, d.Version(stats.Version), nil
}

const releasePreservedSeatsScript = `
    local stats_key = KEYS[1]
    local seat_cnt = ARGV[1]
    redis.call('HINCRBY', stats_key, "Preserved", seat_cnt)
    redis.call('HINCRBY', stats_key, "Version", 1)
    return "ok"
`

func (r *RedisHostDeskRepository) ReleasePreservedSeats(ctx context.Context, partyID d.PartyID) error {
	partyStateKey := r.keys.getPartyStateKey(partyID)
	if exists := r.client.Exists(ctx, partyStateKey).Val(); exists == 0 {
		return domain.ErrPartyNotFound
	}

	results, err := r.client.HMGet(ctx, partyStateKey, "Status", "SeatsCount").Result()
	if err != nil {
		return err
	}

	if results[0].(string) != string(domain.SeatPreserved) {
		return domain.ErrPartyNoPreservedSeats
	}

	seats := results[1].(int64)
	script := redis.NewScript(releasePreservedSeatsScript)
	_, err = script.Run(ctx, r.client, []string{r.keys.getStatsKey()}, int(seats)).Result()
	if err != nil {
		return err
	}

	r.logger.LogDebug(REDIS_HOSTDESK, "release preserved seats", "party id", partyID, "seat count", seats)
	return nil
}

const transferToOccupiedScript = `
    local stats_key = KEYS[1]
    local party_state_key = KEYS[2]
    local seat_cnt = ARGV[1]
    local party_next_status = ARGV[2]
    local checked_in_at = ARGV[3]
    redis.call('HINCRBY', stats_key, 'Occupied', seat_cnt)
    redis.call('HINCRBY', stats_key, 'Preserved', -seat_cnt)
    redis.call('HINCRBY', stats_key, 'Version', 1)
    redis.call('HMSET', party_state_key, "Status", party_next_status, "CheckedInAt", checked_in_at)
    return nil
`

func (r *RedisHostDeskRepository) TransferToOccupied(ctx context.Context, partyID d.PartyID) error {
	partyStateKey := r.keys.getPartyStateKey(partyID)
	if exists := r.client.Exists(ctx, partyStateKey).Val(); exists == 0 {
		return domain.ErrPartyNotFound
	}

	results, err := r.client.HMGet(ctx, partyStateKey, "Status", "SeatsCount").Result()
	if err != nil {
		return err
	}

	if results[0].(string) != string(domain.SeatPreserved) {
		return domain.ErrPartyNoPreservedSeats
	}

	seats, ok := results[1].(int64)
	if !ok {
		seats, _ = strconv.ParseInt(results[1].(string), 10, 64)
	}
	script := redis.NewScript(transferToOccupiedScript)
	transferKeys := []string{r.keys.getStatsKey(), r.keys.getPartyStateKey(partyID)}
	checkedInAt := time.Now().UTC()
	transferVals := []interface{}{seats, domain.SeatOccupied, checkedInAt}
	_, err = script.Run(ctx, r.client, transferKeys, transferVals...).Result()

	if err != nil && err != redis.Nil {
		return err
	}

	r.logger.LogDebug(REDIS_HOSTDESK, "transfer to occupied",
		"party id", partyID,
		"party status", domain.SeatOccupied,
		"seat count", seats,
	)
	return nil
}

func (r *RedisHostDeskRepository) GetPartyServiceState(ctx context.Context, partyID d.PartyID) (*domain.PartyServiceState, error) {
	res := r.client.HGetAll(ctx, r.keys.getPartyStateKey(partyID))
	r.logger.LogDebug(REDIS_HOSTDESK, "get party service state", "party", res.Val(), "key", r.keys.getPartyStateKey(partyID))
	if res.Err() != nil {
		return nil, res.Err()
	}

	if len(res.Val()) == 0 {
		return nil, nil
	}

	state := &domain.PartyServiceState{}
	if err := res.Scan(state); err != nil {
		r.logger.LogErr(REDIS_HOSTDESK, err, "could not parse party service state")
		return nil, err
	}

	return state, nil
}

func (r *RedisHostDeskRepository) CreatePartyServiceState(ctx context.Context, state *domain.PartyServiceState) error {
	return r.OptimisticCreatePartyServiceState(ctx, state, d.Version(SKIP_VERSION_CHECK))
}

const createPartyScript = `
    local stats_key = KEYS[1]
    local party_state_key = KEYS[2]
    local version = ARGV[1]             -- -1 -> skip version check
    local seat_in_used_type = ARGV[2]
    local party_id = ARGV[3]            -- ID           domain.PartyID
    local seat_status = ARGV[4]         -- Status       SeatStatus
    local seat_cnt = ARGV[5]            -- SeatsCount   int
    local time = ARGV[6]                -- PreservedAt/CheckedInAt  time.Time

    if tonumber(version) ~= -1 then
        local current_version = redis.call("HGET", stats_key, "Version") or 0
        if tonumber(current_version) ~= tonumber(version) then
            return redis.error_reply("ErrVersionMismatch")
        end
    end

    redis.call('HINCRBY', stats_key, seat_in_used_type, seat_cnt)
    redis.call('HINCRBY', stats_key, 'Version', 1)
    
    local time_field = "CheckedInAt"
    if seat_in_used_type == "Preserved" then
        time_field = "PreservedAt"
    end
    redis.call('HMSET', party_state_key, "ID", party_id, "Status", seat_status, "SeatsCount", seat_cnt, time_field, time)
    return nil
`

func (r *RedisHostDeskRepository) OptimisticCreatePartyServiceState(ctx context.Context, state *domain.PartyServiceState, version d.Version) error {
	partyStateKey := r.keys.getPartyStateKey(state.ID)

	if exists := r.client.Exists(ctx, partyStateKey).Val(); exists == 1 {
		return domain.ErrPartyAlreadyExists
	}

	var seatInUsedType string
	if state.Status == domain.SeatPreserved {
		seatInUsedType = "Preserved"
	} else if state.Status == domain.SeatOccupied {
		seatInUsedType = "Occupied"
	} else {
		return fmt.Errorf("unknown seat status while creating party service state: %v", state.Status)
	}

	script := redis.NewScript(createPartyScript)
	createKeys := []string{r.keys.getStatsKey(), r.keys.getPartyStateKey(state.ID)}
	createVals := []interface{}{
		int(version),
		seatInUsedType,
		string(state.ID),
		string(state.Status),
		state.SeatsCount,
		time.Now().UTC(),
	}
	_, err := script.Run(ctx, r.client, createKeys, createVals...).Result()
	if err != nil && err != redis.Nil {
		if err.Error() == "ERR ErrVersionMismatch" {
			return d.ErrVersionMismatch
		}
		return err
	}

	r.logger.LogDebug(REDIS_HOSTDESK, "create party service state",
		"party id", state.ID,
		"status", state.Status,
		"seats", state.SeatsCount,
	)
	return nil
}

func (r *RedisHostDeskRepository) UpdatePartyServiceState(ctx context.Context, partyID d.PartyID, nextState *domain.PartyServiceState) error {
	state, err := r.GetPartyServiceState(ctx, partyID)
	if err != nil {
		return err
	}

	if err := copier.CopyWithOption(state, nextState, copier.Option{IgnoreEmpty: true}); err != nil {
		return err
	}

	// TODO: not support seat count changes yet
	// if you have to support seat count change, then we need to use lua script to run commands for atomic operations

	return nil
}

const endOfPartyServiceScript = `
    local stats_key = KEYS[1]
    local party_state_key = KEYS[2]
    local seats = redis.call('HGET', party_state_key, "SeatsCount")
    if not seats then
        return 0
    end
    redis.call('HINCRBY', stats_key, "Occupied", -tonumber(seats))
    redis.call('HINCRBY', stats_key, "Version", 1)
    return redis.call('DEL', party_state_key)
`

func (r *RedisHostDeskRepository) EndPartyServiceState(ctx context.Context, partyID d.PartyID) error {
	endOfServiceKeys := []string{r.keys.getStatsKey(), r.keys.getPartyStateKey(partyID)}
	success, err := r.client.Eval(ctx, endOfPartyServiceScript, endOfServiceKeys).Result()

	if err != nil {
		r.logger.LogErr(REDIS_HOSTDESK, err, "could not execute end of party service script on redis", "keys", endOfServiceKeys)
		return err
	}

	if success.(int64) == int64(0) {
		return domain.ErrPartyNotFound
	}

	return nil
}
