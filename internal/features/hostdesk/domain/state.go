package domain

import (
	"queue-bite/internal/domain"
	"time"
)

type SeatStatus string

func (id SeatStatus) MarshalBinary() ([]byte, error) {
	return []byte(string(id)), nil
}

func (id *SeatStatus) UnmarshalBinary(data []byte) error {
	*id = SeatStatus(data)
	return nil
}

const (
	SeatAvailable SeatStatus = "available"
	SeatPreserved SeatStatus = "preserved"
	SeatOccupied  SeatStatus = "occupied"
)

type PartyServiceState struct {
	ID           domain.PartyID `redis:"ID"`
	Status       SeatStatus     `redis:"Status"`
	SeatsCount   int            `redis:"SeatsCount"`
	PreservedAt  time.Time      `redis:"PreservedAt"`
	CheckedInAt  time.Time      `redis:"CheckedInAt"`
	ServiceEndAt time.Time      `redis:"ServiceEndAt"`
}

func NewPartyServiceFromPreserve(partyID domain.PartyID, seats int) *PartyServiceState {
	return &PartyServiceState{
		ID:          partyID,
		Status:      SeatPreserved,
		SeatsCount:  seats,
		PreservedAt: time.Now().UTC(),
	}
}

func NewPartyServiceImmediately(partyID domain.PartyID, seats int) *PartyServiceState {
	return &PartyServiceState{
		ID:          partyID,
		Status:      SeatPreserved,
		SeatsCount:  seats,
		PreservedAt: time.Now().UTC(),
		CheckedInAt: time.Now().UTC(),
	}
}
