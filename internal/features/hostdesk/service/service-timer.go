package service

import (
	"context"
	"sync"
	"time"

	log "queue-bite/internal/config/logger"
	"queue-bite/internal/domain"
	wld "queue-bite/internal/features/waitlist/domain"
)

type ServiceCompletionCallback func(ctx context.Context, partyID domain.PartyID) error

type ServiceTimer interface {
	StartTracking(ctx context.Context, partyID *wld.QueuedParty, onComplete ServiceCompletionCallback) error
}

type linearServiceTimer struct {
	logger           log.Logger
	timers           map[domain.PartyID]*time.Timer
	durationPerGuest time.Duration
	mu               sync.Mutex
}

func NewLinearServiceTimer(logger log.Logger, durationPerGuest time.Duration) ServiceTimer {
	return &linearServiceTimer{
		logger:           logger,
		timers:           make(map[domain.PartyID]*time.Timer),
		durationPerGuest: durationPerGuest,
		mu:               sync.Mutex{},
	}
}

func (t *linearServiceTimer) StartTracking(ctx context.Context, party *wld.QueuedParty, onComplete ServiceCompletionCallback) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	period := t.durationPerGuest * time.Duration(party.Size)
	t.timers[party.ID] = time.AfterFunc(period, func() {
		if err := onComplete(context.Background(), party.ID); err != nil {
			t.logger.LogErr("servicetimer/linear", err, "failed on timer completed", "duration", period, "party", party)
		}
		t.mu.Lock()
		delete(t.timers, party.ID)
		defer t.mu.Unlock()
		t.logger.LogDebug("servicetimer/linear", "end of service timer", "duration", period, "party", party)
	})
	t.logger.LogDebug("servicetimer/linear", "start service timer for party checkin", "duration", period, "party", party)

	return nil
}
