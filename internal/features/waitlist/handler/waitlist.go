package handler

import (
	"queue-bite/internal/features/waitlist/service"
)

type WaitlistHandler struct {
	waitlist service.Waitlist
}

func NewWaitlistHandler(waitlist service.Waitlist) *WaitlistHandler {
	return &WaitlistHandler{
		waitlist: waitlist,
	}
}
