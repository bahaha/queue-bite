package domain

import "errors"

var (
	ErrInsufficientCapacity = errors.New("insufficient seating capacity")
)

var (
	ErrPartyAlreadyExists    = errors.New("party already exists in host desk")
	ErrPartyNotFound         = errors.New("party not found in seats")
	ErrPartyNoPreservedSeats = errors.New("party has no preserved seats")
	ErrPartyAlreadyReady     = errors.New("party is already ready")
	ErrPartyAlreadySeated    = errors.New("party is already seated")
)
