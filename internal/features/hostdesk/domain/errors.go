package domain

import "errors"

var (
	ErrPartyNotFound      = errors.New("party not found in seats")
	ErrPartyAlreadySeated = errors.New("party is already seated")
)
