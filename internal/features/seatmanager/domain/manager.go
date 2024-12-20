package domain

import "context"

type SeatManager interface {
	WatchSeatVacancy(ctx context.Context) error

	UnwatchSeatVacancy(ctx context.Context) error

	HandleCapacityAvailable(ctx context.Context, event interface{}) error

	HandleServiceCompleted(ctx context.Context, event interface{}) error
}
