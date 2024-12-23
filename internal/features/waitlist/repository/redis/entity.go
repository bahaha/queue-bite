package redis

import (
	d "queue-bite/internal/domain"
	"queue-bite/internal/features/waitlist/domain"
	"time"

	"github.com/jinzhu/copier"
)

type redisQueuedParty struct {
	// Flattened fields from Party domain with Redis tags
	ID       string        `redis:"id"`
	Name     string        `redis:"name"`
	Size     int           `redis:"size"`
	JoinedAt time.Time     `redis:"joined_at"`
	Status   d.PartyStatus `redis:"status"`

	// Queue-specific fields
	Position             int `redis:"-"` // Computed from ZRANK
	EstimatedServiceTime int `redis:"est"`
}

func (r *redisQueuedParty) asQueuedParty() *domain.QueuedParty {
	party := &domain.QueuedParty{}
	copier.Copy(party, r)
	party.EstimatedServiceTime = time.Duration(r.EstimatedServiceTime) * time.Second
	return party
}

func newRedisQueuedParty(party *domain.QueuedParty) *redisQueuedParty {
	entity := &redisQueuedParty{}
	copier.Copy(entity, party)
	entity.EstimatedServiceTime = int(party.EstimatedServiceTime.Seconds())
	return entity
}
