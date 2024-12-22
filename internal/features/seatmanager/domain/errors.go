package domain

import "errors"

var (
	ErrPreserveSeats = errors.New("failed to preserve seats")
	ErrJoinWaitlist  = errors.New("failed to join waitlist")
)
