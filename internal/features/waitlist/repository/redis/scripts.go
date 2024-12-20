package redis

// KEYS:
//   - waitlist_key
//   - party_detail_key
//   - total_wait_prefixsum_key
//   - party_wait_prefixsum_key
//
// ARGV:
//   - party_id
//   - estimated_service_time
//   - join_score
//   - ttl
const joinScript = `
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
local success_cnt = redis.call('ZADD', waitlist_key, 'NX', join_score, party_id)
redis.call('EXPIRE', waitlist_key, ttl)

if success_cnt == 0 then
    return {success_cnt}
end

-- Increment total wait time and get new value
local next_wait = redis.call('INCRBY', total_wait_prefixsum_key, estimated_service_time)

-- Set wait time prefix sum for this group with TTL
redis.call('SET', party_wait_prefixsum_key, next_wait, 'EX', ttl)

return {success_cnt, wait_entries_ahead, next_wait}
`

// KEYS:
//   - party_detail_key
//   - waitlist_key
//   - party_wait_prefixsum_key
//
// ARGV:
//   - party_id
const getPartyScript = `
local party_detail_key = KEYS[1]
local waitlist_key = KEYS[2]
local party_wait_prefixsum_key = KEYS[3]
local total_service_time_key = KEYS[4]
local party_id = ARGV[1]

local details = redis.call('HGETALL', party_detail_key)
if #details == 0 then
    return nil
end

local wait_time = redis.call('GET', party_wait_prefixsum_key)
if not wait_time then
    return nil
end

local total_service_time = redis.call('GET', total_service_time_key) or 0

local position = redis.call('ZRANK', waitlist_key, party_id)
if not position then
    return nil
end

return {details, wait_time - total_service_time, position}
`

// KEYS:
//   - waitlist_key
//   - party_detail_key
//   - total_service_time_key
//   - party_wait_prefixsum_key_prefix
//
// ARGV:
//   - party_id
//   - estimated_service_time_field
const leaveScript = `
local waitlist_key = KEYS[1]
local party_detail_key = KEYS[2]
local total_service_time_key = KEYS[3]
local party_wait_prefixsum_key_prefix = KEYS[4] 
local party_id = ARGV[1]
local estimated_service_time_field = ARGV[2]

local rank = redis.call('ZRANK', waitlist_key, party_id)
if not rank then
    return nil
end

local est = tonumber(redis.call('HGET', party_detail_key, estimated_service_time_field))
if not est then
-- TODO: find an approach to handle the state inconsistency
    return nil
end

if rank == 0 then
    redis.call('INCRBY', total_service_time_key, est)
else
    local affected = redis.call('ZRANGE', waitlist_key, rank + 1, -1)
    -- TODO: find an approach to handle update on too much queued entity
    for _, party in ipairs(affected) do
        local prefixsum_key = party_wait_prefixsum_key_prefix .. party
        redis.call('INCRBY', prefixsum_key, -est)
    end
end

redis.call('ZREM', waitlist_key, party_id)
return {rank, est}
`
