package domain

import (
	"queue-bite/internal/domain"
	"time"
)

// QueuedParty represents a party currently waiting in the queue.
// It extends the base Party type with queue-specific information.
type QueuedParty struct {
	*domain.Party
	// Position in the queue (0-based). Lower number indicates earlier position.
	Position int
	// Total time this party expects to wait before being served.
	EstimatedEndOfServiceTime time.Duration
	JoinedAt                  time.Time
}

func (p *QueuedParty) RemainingWaitTime() time.Duration {
	return p.EstimatedEndOfServiceTime - p.EstimatedServiceTime
}

// QueueStatus provides information about the current state of the waitlist.
type QueueStatus struct {
	TotalParties int
	// Estimated wait time for a new party joining now
	CurrentWaitTime time.Duration
}
