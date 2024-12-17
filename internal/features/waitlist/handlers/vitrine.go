package waitlist

import (
	"net/http"

	"github.com/a-h/templ"

	view "queue-bite/internal/features/waitlist/views"
)

type waitlistVitrineHandler struct{}

func newWaitlistVitrineHandler() *waitlistVitrineHandler {
	return &waitlistVitrineHandler{}
}

func (h *waitlistVitrineHandler) GetVitrineDisplay() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(view.Vitrine(&view.VitrinePageData{
			Form: view.NewJoinFormData(),
		})).ServeHTTP(w, r)
	}
}
