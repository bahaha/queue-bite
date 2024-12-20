package repository

import "context"

type HostDeskRepository interface {
	GetCurrentOccupiedSeats(ctx context.Context) (int, error)
}
