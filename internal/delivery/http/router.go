package http

import (
	"net/http"
	v1 "subscription-api/internal/delivery/http/v1"
	"subscription-api/internal/domain"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(uc domain.SubscriptionUseCase) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromHeader("X-Real-IP"))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	handler := v1.NewSubscriptionHandler(uc)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/subscriptions", func(r chi.Router) {
			r.Post("/", handler.Create)
			r.Get("/total", handler.CalculateTotal)
			r.Get("/{id}", handler.Get)
			r.Put("/{id}", handler.Update)
			r.Delete("/{id}", handler.Delete)
		})
	})

	return r
}
