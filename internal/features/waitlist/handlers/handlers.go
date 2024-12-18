package waitlist

import (
	log "queue-bite/internal/config/logger"
	"queue-bite/internal/platform"
)

const (
	WAITLIST = "/waitlist"
)

type WaitlistHandlers struct {
	Waitlist *waitlistHandler
	Vitrine  *waitlistVitrineHandler
}

func NewWaitlistHandlers(logger log.Logger, redis *platform.RedisComponent) *WaitlistHandlers {
	return &WaitlistHandlers{
		Waitlist: newWaitlistHandler(logger, redis.Client),
		Vitrine:  newWaitlistVitrineHandler(),
	}
}
