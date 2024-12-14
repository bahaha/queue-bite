package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	waitlist "queue-bite/internal/features/waitlist/handlers"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Handle("/assets/*", http.FileServer(http.FS(Files)))

	r.Get("/", redirect("/waitlist", http.StatusTemporaryRedirect))

	r.Route("/waitlist", func(r chi.Router) {
		handler := waitlist.NewWaitlistHandlers()
		r.Get("/", handler.Vitrine.GetVitrineDisplay)
	})

	return r
}

func redirect(path string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, path, status)
	}
}
