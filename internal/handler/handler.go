package handler

import (
	"encoding/json"
	"net/http"

	"github.com/onbehalfofhim/gofermart/internal/models"
	"github.com/onbehalfofhim/gofermart/internal/service"
)

// HTTP хендлер для приложения
type Handler struct {
	userService *service.UserService
}

// констуктор для HTTP хендлера
func NewHandler(u *service.UserService) *Handler {
	return &Handler{
		userService: u,
	}
}

// обработчик регистрации нового пользователя
func (h *Handler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var req models.AuthRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.Login == "" || req.Password == "" {
			http.Error(w, "login and password required", http.StatusBadRequest)
			return
		}

		token, err := h.userService.Register(req.Login, req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		resp := models.AuthResponse{
			Token: token,
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// обработчик аутентификации пользователя
func (h *Handler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var req models.AuthRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.Login == "" || req.Password == "" {
			http.Error(w, "login and password required", http.StatusBadRequest)
			return
		}

		token, err := h.userService.Login(req.Login, req.Password)
		if err != nil {
			http.Error(w, "invalid login/password", http.StatusUnauthorized)
			return
		}

		resp := models.AuthResponse{
			Token: token,
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
