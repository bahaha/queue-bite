package utils

import "time"

type Clock struct {
	Fixed time.Time
}

func (c Clock) Now() time.Time {
	if c.Fixed.IsZero() {
		return time.Now()
	}
	return c.Fixed
}
