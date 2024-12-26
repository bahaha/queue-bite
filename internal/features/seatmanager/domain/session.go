package domain

import d "queue-bite/internal/domain"

type PartySession struct {
	ID   d.PartyID
	Name string
	Size int
}
