package domain

import "time"

type PartyID string

func (id PartyID) MarshalBinary() ([]byte, error) {
	return []byte(string(id)), nil
}

func (id *PartyID) UnmarshalBinary(data []byte) error {
	*id = PartyID(data)
	return nil
}

type Party struct {
	ID   PartyID
	Name string
	Size int
	// Estimated time needed to serve this party once seated.
	EstimatedServiceTime time.Duration
}

func NewParty(name string, size int) *Party {
	return &Party{
		Name: name,
		Size: size,
	}
}
