package handler

import (
	"net/http"
	d "queue-bite/internal/domain"
)

type seatManagerHandler struct{}

func NewSeatManagerHandler() *seatManagerHandler {
	return &seatManagerHandler{}
}

type PartySession struct {
	PartyID d.PartyID
}

func redirectToVisitPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
