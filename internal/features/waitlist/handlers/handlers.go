package waitlist

type WaitlistHandler struct {
	Vitrine *WaitlistVitrineHandler
}

func NewWaitlistHandlers() *WaitlistHandler {
	return &WaitlistHandler{
		Vitrine: newWaitlistVitrineHandler(),
	}
}
