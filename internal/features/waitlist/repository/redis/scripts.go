package redis

// joinScript atomically adds a party to the waitlist queue and updates timing metrics
//
// Keys:
//
//	waitlist_key           - Sorted set of parties in queue order
//	party_detail_key       - Hash storing party details
//	total_wait_prefixsum   - Total wait time for all parties
//	party_wait_prefixsum   - Individual party's wait time prefixsum
//	total_service_time     - Total service time of all parties
//	waiting_party_counter  - Number of parties in waiting status
//
// Args:
//
//	party_id              - Unique party identifier
//	estimated_service_time - Expected service duration in seconds
//	join_score            - Score for queue ordering (timestamp)
//	ttl                   - TTL for keys in seconds
//	is_party_waiting      - "1" if party starts in waiting status
//
// Returns: [success_cnt, position, wait_time]
//
//	success_cnt = 0: Party already exists
//	success_cnt = 1: Successfully added
//	position: Party's position in queue (0-based)
//	wait_time: Estimated wait time for this party
const joinScript = `
local waitlist_key = KEYS[1]
local party_detail_key = KEYS[2]
local total_wait_prefixsum_key = KEYS[3]
local party_wait_prefixsum_key = KEYS[4]
local total_service_time_key = KEYS[5]
local waiting_party_counter_key = KEYS[6]
local party_id = ARGV[1]
local estimated_service_time = ARGV[2]
local join_score = ARGV[3]
local ttl = ARGV[4]
local is_party_waiting = ARGV[5]

-- Add to sorted set
local wait_entries_ahead = redis.call('ZCARD', waitlist_key)
local success_cnt = redis.call('ZADD', waitlist_key, 'NX', join_score, party_id)
redis.call('EXPIRE', waitlist_key, ttl)

if success_cnt == 0 then
    return {success_cnt}
end

-- Increment total wait time and get new value
local next_wait = redis.call('INCRBY', total_wait_prefixsum_key, estimated_service_time)
local total_service_time = redis.call('GET', total_service_time_key) or 0

-- Set wait time prefix sum for this group with TTL
redis.call('SET', party_wait_prefixsum_key, next_wait, 'EX', ttl)

if is_party_waiting == "1" then
    redis.call('INCR', waiting_party_counter_key)
end

return {success_cnt, wait_entries_ahead, next_wait - total_service_time}
`

// getPartyScript atomically retrieves party details and queue position
//
// Keys:
//
//	party_detail_key     - Hash storing party details
//	waitlist_key         - Queue ordered set
//	party_wait_prefixsum - Party's wait time
//	total_service_time   - Total active service time
//
// Args:
//
//	party_id             - Party to retrieve
//
// Returns: [details, wait_time, position] or nil if party not found
//
//	details: Hash of party details
//	wait_time: Current wait time estimate
//	position: Queue position (0-based)
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

// leaveScript atomically removes party and updates queue metrics
//
// Keys:
//
//	waitlist_key              - Queue ordered set
//	party_detail_key          - Party details hash
//	total_service_time        - Service time counter
//	party_wait_prefixsum_prefix - Prefix for wait time keys
//	total_wait_prefixsum      - Total wait counter
//	waiting_party_counter     - Waiting status counter
//
// Args:
//
//	party_id                  - Party to remove
//	estimated_service_time_field - Field name for service time
//	status_field             - Field name for status
//	status_party_wait_val    - Status value for waiting
//
// Returns: [position, service_time] or nil if not found
//
//	position: Party's position in queue
//	service_time: Party's service duration
//	has_queue: Whether queue still exists
//	deleted_keys: Keys that were cleaned up
const leaveScript = `
local waitlist_key = KEYS[1]
local party_detail_key = KEYS[2]
local total_service_time_key = KEYS[3]
local party_wait_prefixsum_key_prefix = KEYS[4] 
local total_wait_prefixsum_key = KEYS[5]
local waiting_party_counter_key = KEYS[6]
local party_id = ARGV[1]
local estimated_service_time_field = ARGV[2]
local status_field = ARGV[3]
local status_party_wait_val = ARGV[4]

local rank = redis.call('ZRANK', waitlist_key, party_id)
if not rank then
    return nil
end

local party = redis.call('HMGET', party_detail_key, estimated_service_time_field, status_field)
local est = tonumber(party[1])
if not est then
-- TODO: find an approach to handle the state inconsistency
    return nil
end

local status = party[2]
if status == status_party_wait_val then
    redis.call('INCRBY', waiting_party_counter_key, -1)
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
local del_keys = {party_detail_key, party_wait_prefixsum_key_prefix .. party_id}
local has_wait = redis.call('EXISTS', waitlist_key)
if has_wait == 0 then
    table.insert(del_keys, waitlist_key)
    table.insert(del_keys, total_service_time_key)
    table.insert(del_keys, total_wait_prefixsum_key)
end
redis.call('DEL', unpack(del_keys))
return {rank, est, has_wait}
`
