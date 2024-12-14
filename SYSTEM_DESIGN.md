# Remote Waitlist Manager - System Design

## Requirements

### Functional Requirements

1. Waitlist Join Process
  - New diners can submit their party details (name and size)
  - System validates and processes join requests
  - Parties receive confirmation of their position
  
2. Real-time Status Updates
  - Parties can view their current position without refreshing
  - System pushes updates for queue position changes
  - Notifies parties when their table is ready

3. Check-in Process
  - Eligible parties see a check-in button when table is ready
  - System updates available seats upon check-in
  - Service timer starts automatically after check-in

4. Session Management
  - Parties can view their status across different browser sessions
  - Status persists event if browser is closed and reopened
  - Secure party identification through sessions

### Non-Functional Requirements

1. Queue Management
  - Display estimated wait times for queued parties
  - Show number of parties ahead in line
  - Support different seating preferences (counter vs tables)

2. Restaurant Operations
  - Configurable queue management strategies
  - Flexible seating allocation policies
  - Check-in timeout handling

3. User Experience
  - Proactive notifications before service completion
  - Easy process to leave the waitlist
  - Clear status indicators throughout the process


## Architecture

### System Components

1. Queue Management System
  - Redis-based queue implementation
  - Optimized list structure for party management
  - Real-time state management

2. Communication Layer
  - Server-Sent Events (SSE) for real-time updates
  - HTMX for dynamic UI updates
  - Session-based party identification

#### Why Redis?

##### Data Structure Advantages
1. List Operations
  - Native support for queue operations (LPUSH, RPOP)
  - Atomic operations ensure data consistency
  - Up to 2^32 elements (~4.2 billion) per list
  - Listpack encoding for small lists (<512 elements) optimizes memory
  - O(1) complexity for queue operations
2. Key Design
```
wl:{restaurant_id}:tq       # Ordered list of waiting parties (queue for tables)
wl:{restaurant_id}:cq       # Ordered list of waiting parties (queue for counter)
wl:{restaurant_id}:active   # Hash of active dining parties
party:{id}                  # Hash storing party details
```

##### Scalability

1. Distributed System Support
  - Horizontal scaling of web servers
  - Simple PUB/SUB event broadcasting support for simplicty without introduce Message-Queue structure.
  ```
    [Redis] ---> [Pub/Sub] ---> [Server 1] ---> [SSE Clients]
                           ---> [Server 2] ---> [SSE Clients]
                           ---> [Server 3] ---> [SSE Clients]
  ```
  - Potential distribution lock solution
  - Redis Cluster for larger deployments

2. Concurrent Operations
  - Atomic list operations prevent race conditions
  - Lua scripts for custom atomic operations / transaction support

##### Reliability

- Redis Sentinel or Redis Cluster for automatic failover
- RDB snapshots for point-in-time recovery with AOF logs for operation replay

##### Performance Optimization

1. Memory Efficiency
  - Listpack Encoding optimization for small lists (<512 elements) with reduced memory overhead and better CPU cache utilization which covers 90%+ restaurants

2. Operations Optimization
  - Lua scripts for complex operations while keep atomic
  - Replica nodes for read scaling
  - Cache frequently accessed data

### Data Model
1. Restaurant Entity
```
- Total capacity
- Seating configurations
- Check-in policy
- Queue management strategies
```

2. Party Entity
```
- Unique identifier
- Party name
- Party size
- Current state (pending/waiting/ready/dining)
- Timestamps (join time, update time, complete time)
```

3. Table Entity
```
- Table number
- Capacity
- Current state (ready/occupied/cleaning)
- Current party (if occupied)
- Type (counter/table)
```

## Interface

#### Waitlist management
```
GET  /waitlist/{restaurant_id}          # Display waitlist and join form
POST /waitlist/{restaurant_id}/join     # Submit join request
POST /waitlist/{restaurant_id}/checkin  # Party check-in
```

#### Real-time Updates
```
GET /waitlist/{restaurant_id}/events    # SSE endpoint for updates
```

## Optimization Considerations
- Ready heavy optimization strategies
- Handle concurrent join requests efficiently
- Consistent state management
