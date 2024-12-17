package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	waitlist "queue-bite/internal/features/waitlist/handlers"
	"queue-bite/internal/platform"
	"queue-bite/pkg/utils"
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
	r.Get("/healthz", healthHandler(s.redis))

	r.Route("/waitlist", func(r chi.Router) {
		handler := waitlist.NewWaitlistHandlers()
		r.Get("/", handler.Vitrine.GetVitrineDisplay())
		r.Post("/join", handler.Waitlist.JoinWaitlist(s.logger, s.validate, s.translators))
	})

	return r
}

func redirect(path string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, path, status)
	}
}

func healthHandler(systemComponents ...platform.SystemComponents) http.HandlerFunc {
	components := make(map[string]map[string]string)

	return func(w http.ResponseWriter, r *http.Request) {
		for _, comp := range systemComponents {
			components[comp.Name()] = comp.Health()
		}

		err := utils.Encode(w, r, http.StatusOK, components)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
