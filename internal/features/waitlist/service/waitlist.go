package service

import (
	"context"
	"time"

	"github.com/jinzhu/copier"

	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	servicetime "queue-bite/internal/features/servicetime/service"
	"queue-bite/internal/features/sse"
	"queue-bite/internal/features/waitlist/domain"
	"queue-bite/internal/features/waitlist/repository"
	"queue-bite/internal/platform/eventbus"
)

var WAITLIST = "waitlist"

type QueuedPartyProvider interface {

	// GetQueuedParties returns a channel that streams parties currently in the queue.
	// Parties are yielded in their queue order (FIFO).
	// The channel is closed when all parties have been streamed or context is cancelled.
	// This method allows for efficient iteration over large queues.
	GetQueuedParties(ctx context.Context) (<-chan *domain.QueuedParty, error)
}

// Waitlist provides operations for managing a waiting queue.
// It handles party joining, leaving, and status queries while maintaining
// wait time calculations and queue positions.
type Waitlist interface {
	QueuedPartyProvider

	// HasPartyExists return if a party in waitlist queue.
	HasPartyExists(ctx context.Context, partyID d.PartyID) bool

	// JoinQueue adds a new party to the waitlist queue.
	// It calculates estimated service duration and waiting time,
	// assigns a queue position, and returns the queued party information.
	// Returns ErrPartyAlreadyQueued if the party is already in queue.
	JoinQueue(ctx context.Context, party *d.Party) (*domain.QueuedParty, error)

	// LeaveQueue removes a party from the waitlist queue.
	// This updates wait times for remaining parties in the queue.
	// If the party is not found, it returns nil without error.
	LeaveQueue(ctx context.Context, partyID d.PartyID) error

	// GetQueueStatus returns the current state of the waitlist,
	// including total number of waiting parties and estimated wait time
	// for new parties joining the queue.
	GetQueueStatus(ctx context.Context) (*domain.QueueStatus, error)

	// GetQueuedParty retrieves the current status of a party in the queue.
	// Returns the party's current position and estimated waiting time.
	// Returns nil, nil if the party is not found in the queue.
	GetQueuedParty(ctx context.Context, partyID d.PartyID) (*domain.QueuedParty, error)

	// HandlePartyReady handle the event of party ready
	// if the queued party status is not waiting, throws ErrInvalidStateTransition
	HandlePartyReady(ctx context.Context, partyID d.PartyID) error
}

type waitlistService struct {
	logger           log.Logger
	repo             repository.WaitlistRepositoy
	eventbus         eventbus.EventBus
	serviceEstimator servicetime.ServiceTimeEstimator
}

func NewWaitlistService(
	logger log.Logger,
	repo repository.WaitlistRepositoy,
	estimator servicetime.ServiceTimeEstimator,
	eventbus eventbus.EventBus,
) Waitlist {
	return &waitlistService{
		logger:           logger,
		repo:             repo,
		eventbus:         eventbus,
		serviceEstimator: estimator,
	}
}

func (s *waitlistService) HasPartyExists(ctx context.Context, partyID d.PartyID) bool {
	return s.repo.HasParty(ctx, partyID)
}

func (s *waitlistService) JoinQueue(ctx context.Context, party *d.Party) (*domain.QueuedParty, error) {
	serviceDuration, err := s.serviceEstimator.EstimateServiceTime(ctx, party)
	if err != nil {
		return nil, err
	}
	party.EstimatedServiceTime = serviceDuration.Duration

	queuedParty := &domain.QueuedParty{}
	copier.Copy(queuedParty, party)
	queuedParty.JoinedAt = time.Now()

	queuedParty, err = s.repo.AddParty(ctx, queuedParty)
	if err != nil {
		return nil, err
	}

	return queuedParty, nil
}

func (s *waitlistService) LeaveQueue(ctx context.Context, partyID d.PartyID) error {
	return s.repo.RemoveParty(ctx, partyID)
}

func (s *waitlistService) GetQueueStatus(ctx context.Context) (*domain.QueueStatus, error) {
	return s.repo.GetQueueStatus(ctx)
}

func (s *waitlistService) GetQueuedParty(ctx context.Context, partyID d.PartyID) (*domain.QueuedParty, error) {
	return s.repo.GetParty(ctx, partyID)
}

func (s *waitlistService) GetQueuedParties(ctx context.Context) (<-chan *domain.QueuedParty, error) {
	return s.repo.ScanParties(ctx)
}

func (s *waitlistService) HandlePartyReady(ctx context.Context, partyID d.PartyID) error {
	queuedParty, err := s.repo.GetPartyDetails(ctx, partyID)
	if err != nil {
		return err
	}

	if queuedParty == nil {
		s.logger.LogDebug("waitlist", "party not found for `ready`", "party id", partyID)
		return domain.ErrPartyNotFound
	}

	if queuedParty.Status == d.PartyStatusReady {
		return nil
	} else if queuedParty.Status != d.PartyStatusWaiting {
		s.logger.LogDebug("waitlist", "invalid party status transition to `ready`", "current status", queuedParty.Status)
		return domain.ErrInvalidPartyStatusTransition
	}

	if err = s.repo.UpdatePartyStatus(ctx, partyID, d.PartyStatusReady); err != nil {
		s.logger.LogErr("waitlist", err, "failed to make party ready", "party id", partyID)
	}

	s.eventbus.Publish(ctx, &sse.NotifyPartyReadyEvent{PartyID: partyID})
	return nil
}
