package domain

import (
	"queue-bite/internal/domain"
	"time"
)

type PartyStatus string

const (
	SeatReady     PartyStatus = "ready"
	PartyNotified PartyStatus = "notified"
	PartySeated   PartyStatus = "seated"
	PartyServing  PartyStatus = "serving"
)

type PartyServiceState struct {
	ID             domain.PartyID
	Status         PartyStatus
	NotifiedAt     time.Time
	ServiceStartAt time.Time
	SeatsOccupied  int
}
