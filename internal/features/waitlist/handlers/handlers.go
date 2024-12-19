package waitlist

import (
	log "queue-bite/internal/config/logger"
	"queue-bite/internal/features/waitlist/services"
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
	estimator := services.NewLinearServiceTimeEstimator()
	service := services.NewRedisWaitlistService(logger, redis.Client, estimator)

	return &WaitlistHandlers{
		Waitlist: newWaitlistHandler(service),
		Vitrine:  newWaitlistVitrineHandler(service),
	}
}
