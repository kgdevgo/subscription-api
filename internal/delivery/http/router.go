package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/kgdevgo/subscription-api/docs"
	v1 "github.com/kgdevgo/subscription-api/internal/delivery/http/v1"
	"github.com/kgdevgo/subscription-api/internal/domain"
)

func NewRouter(uc domain.SubscriptionUseCase) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromHeader("X-Real-IP"))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

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
