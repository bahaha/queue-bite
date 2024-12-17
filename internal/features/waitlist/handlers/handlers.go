package waitlist

const (
	FEAT_WAITLIST = "feat/waitlist"
)

type WaitlistHandlers struct {
	Waitlist *waitlistHandler
	Vitrine  *waitlistVitrineHandler
}

func NewWaitlistHandlers() *WaitlistHandlers {
	return &WaitlistHandlers{
		Waitlist: newWaitlistHandler(),
		Vitrine:  newWaitlistVitrineHandler(),
	}
}
