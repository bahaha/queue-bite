package domain

import (
	"queue-bite/internal/domain"
	"time"
)

type PartyStatus string

const (
	PartyNotified PartyStatus = "notified"
	PartyServing  PartyStatus = "serving"
)

type PartyServiceState struct {
	*domain.Party
	Status         PartyStatus
	NotifiedAt     time.Time
	ServiceStartAt *time.Time
	SeatsOccupied  int
}
