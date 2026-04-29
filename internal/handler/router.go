package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) Route() http.Handler {
	r := chi.NewRouter()

	r.Route("/api/user/", func(r chi.Router) {
		r.Post("/register", h.Register())
		r.Post("/login", h.Login())
	})

	return r
}
