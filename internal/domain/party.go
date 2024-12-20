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

func NewParty(name string, size int) *Party {
	return &Party{
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
	PartyStatusWaiting PartyStatus = "waiting"
	PartyStatusReady   PartyStatus = "ready"
)
