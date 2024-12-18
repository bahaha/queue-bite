package services

import (
	"context"
	"time"

	domain "queue-bite/internal/features/waitlist"
)

type QueuedParty struct {
	domain.Party
	QueuePosition int
	JoinedAt      time.Time
	EstimatedWait time.Duration
}

type QueueStatus struct {
	TotalParties  int
	EstimatedWait time.Duration
}

type WaitlistService interface {
	// GetQueueStatus() (*QueueStatus, error)
	//
	// GetQueuedParty(partyID string) error

	JoinQueue(ctx context.Context, party *domain.Party) (*QueuedParty, error)

	// LeaveQueue(partyID string) error
}

type WaitTimeEstimator interface {
	EstimateWaitTime(ctx context.Context, payty *domain.Party) (time.Duration, error)
}

const (
	WAIT_ESTIMATOR = "waitlist/wait-estimator"
)
