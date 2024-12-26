package handler

import (
	"net/http"
)

type seatManagerHandler struct{}

func NewSeatManagerHandler() *seatManagerHandler {
	return &seatManagerHandler{}
}

func redirectToVisitPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
