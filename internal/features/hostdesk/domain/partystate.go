package domain

import (
	"queue-bite/internal/domain"
	"time"
)

type SeatStatus string

const (
	SeatAvailable SeatStatus = "available"
	SeatPreserved SeatStatus = "preserved"
	SeatOccupied  SeatStatus = "occupied"
)

type PartyServiceState struct {
	ID           domain.PartyID
	Status       SeatStatus
	SeatsCount   int
	PreservedAt  time.Time
	CheckedInAt  time.Time
	ServiceEndAt time.Time
}

func NewPartyServiceFromPreserve(partyID domain.PartyID, seats int) *PartyServiceState {
	return &PartyServiceState{
		ID:          partyID,
		Status:      SeatPreserved,
		SeatsCount:  seats,
		PreservedAt: time.Now(),
	}
}

func NewPartyServiceImmediately(partyID domain.PartyID, seats int) *PartyServiceState {
	return &PartyServiceState{
		ID:          partyID,
		Status:      SeatPreserved,
		SeatsCount:  seats,
		PreservedAt: time.Now(),
		CheckedInAt: time.Now(),
	}
}
