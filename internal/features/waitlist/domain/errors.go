package domain

import (
	"errors"
	"fmt"
	"queue-bite/internal/domain"
)

var (
	ErrPartyAlreadyQueued           = errors.New("party is already in queue")
	ErrPartyNotFound                = errors.New("party not found in queue")
	ErrInvalidPartyStatusTransition = errors.New("invalid party status transition")
)

type QueueOperationError struct {
	Op  string
	ID  domain.PartyID
	Err error
}

func (e *QueueOperationError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("queue operation %s failed for party %s: %v", e.Op, e.ID, e.Err)
	}
	return fmt.Sprintf("queue operation %s failed: %v", e.Op, e.Err)
}
