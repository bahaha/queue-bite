package domain

import (
	"time"
)

type PartyID string

func (id PartyID) MarshalBinary() ([]byte, error) {
	return []byte(string(id)), nil
}

func (id *PartyID) UnmarshalBinary(data []byte) error {
	*id = PartyID(data)
	return nil
}

type Party struct {
	ID     PartyID
	Name   string
	Size   int
	Status PartyStatus
	// Estimated time needed to serve this party once seated.
	EstimatedServiceTime time.Duration
}

func NewParty(id PartyID, name string, size int) *Party {
	return &Party{
		ID:   id,
		Name: name,
		Size: size,
	}
}

type PartyStatus string

func (id PartyStatus) MarshalBinary() ([]byte, error) {
	return []byte(string(id)), nil
}

func (id *PartyStatus) UnmarshalBinary(data []byte) error {
	*id = PartyStatus(data)
	return nil
}

const (
	PartyStatusReady   PartyStatus = "ready"
	PartyStatusWaiting PartyStatus = "waiting"
	PartyStatusServing PartyStatus = "serving"
)
