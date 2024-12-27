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

// QueuedPartyProvider defines interface for streaming parties in queue.
// Useful for iterating through large queues efficiently.
type QueuedPartyProvider interface {
	// GetQueuedParties returns a channel that yields parties in queue order.
	// Channel is closed when iteration completes or context is cancelled.
	GetQueuedParties(ctx context.Context) (<-chan *domain.QueuedParty, error)
}

// Waitlist manages the restaurant's waiting queue operations.
type Waitlist interface {
	QueuedPartyProvider

	HasPartyExists(ctx context.Context, partyID d.PartyID) bool

	JoinQueue(ctx context.Context, party *d.Party) (*domain.QueuedParty, error)

	LeaveQueue(ctx context.Context, partyID d.PartyID) error

	// GetQueueStatus returns current queue metrics like total parties and wait times
	GetQueueStatus(ctx context.Context) (*domain.QueueStatus, error)

	// GetQueuedParty retrieves a specific party's queue information with its position in the queue
	GetQueuedParty(ctx context.Context, partyID d.PartyID) (*domain.QueuedParty, error)

	// HandlePartyReady processes a party becoming ready for seating
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
