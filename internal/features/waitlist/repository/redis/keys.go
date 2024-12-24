package redis

import (
	"fmt"
	"queue-bite/internal/domain"
)

type queueKeys struct{}

// queue:waiting
func (k *queueKeys) waitingQueue() string {
	return "queue:waiting"
}

// queue:party:<id>
func (k *queueKeys) partyDetails(id domain.PartyID) string {
	return fmt.Sprintf("queue:party:%s", id)
}

// queue:waiting:count
func (k *queueKeys) waitingPartyCounter() string {
	return "queue:waiting:count"
}

// "queue:wait:sum"
func (k *queueKeys) waitTimePrefixsum() string {
	return "queue:wait:sum"
}

// queue:wait:<id>
func (k *queueKeys) partyWaitTime(id domain.PartyID) string {
	return fmt.Sprintf("queue:wait:%s", id)
}

// queue:wait:
func (k *queueKeys) partyWaitTimePrefix() string {
	return "queue:wait:"
}

// queue:service
func (k *queueKeys) totalServiceTime() string {
	return "queue:service"
}
