package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onbehalfofhim/gofermart/internal/logger"
	"github.com/onbehalfofhim/gofermart/internal/middleware"
)

// настройка маршрутов для приложения
func (h *Handler) Route(logger *logger.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(logger))

	r.Route("/api/user/", func(r chi.Router) {
		r.Post("/register", h.Register())
		r.Post("/login", h.Login())
	})

	return r
}
