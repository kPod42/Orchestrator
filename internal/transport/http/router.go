package httptransport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"Coordinator/internal/handler"
)

func NewRouter(h *handler.HTTPHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Route("/coordinator", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Get("/nodes", h.GetNodes)
		r.Get("/health", h.Health)
	})

	return r
}
