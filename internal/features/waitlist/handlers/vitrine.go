package waitlist

import (
	"net/http"

	"github.com/a-h/templ"

	view "queue-bite/internal/features/waitlist/views"
)

type WaitlistVitrineHandler struct{}

func newWaitlistVitrineHandler() *WaitlistVitrineHandler {
	return &WaitlistVitrineHandler{}
}

func (h *WaitlistVitrineHandler) GetVitrineDisplay(w http.ResponseWriter, r *http.Request) {
	templ.Handler(view.Virtrine()).ServeHTTP(w, r)
}
