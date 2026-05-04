package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onbehalfofhim/gofermart/internal/logger"
	"github.com/onbehalfofhim/gofermart/internal/middleware"
)

// настройка маршрутов для приложения
func (h *Handler) Route(logger *logger.Logger, jwtManager middleware.JWTValidator) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(logger))

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.Register())
		r.Post("/login", h.Login())

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtManager))
			r.Post("/orders", h.CreateOrder())
			r.Post("/balance/withdraw", h.CreateWithdraw())
		})
	})

	return r
}
