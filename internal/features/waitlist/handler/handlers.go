package handler

import (
	log "queue-bite/internal/config/logger"
	servicetime "queue-bite/internal/features/servicetime/service"
	repo "queue-bite/internal/features/waitlist/repository/redis"
	"queue-bite/internal/features/waitlist/service"
	"queue-bite/internal/platform"
	"time"
)

type WaitlistHandlers struct {
	Vitrine  *VitrineHandler
	Waitlist *WaitlistHandler
}

func NewWaitlistHandlers(logger log.Logger, redis *platform.RedisComponent, estimator servicetime.ServiceTimeEstimator) *WaitlistHandlers {
	redisRepo := repo.NewRedisWaitlistRepository(logger, redis.Client, 24*time.Hour)
	service := service.NewWaitlistService(redisRepo, estimator)

	return &WaitlistHandlers{
		Waitlist: NewWaitlistHandler(service),
		Vitrine:  NewVitrineHandler(service),
	}

}
