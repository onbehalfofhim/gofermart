package handler

import (
	"encoding/json"
	"net/http"

	"github.com/onbehalfofhim/gofermart/internal/logger"
	"github.com/onbehalfofhim/gofermart/internal/models"
	"github.com/onbehalfofhim/gofermart/internal/repository"
	"github.com/onbehalfofhim/gofermart/internal/service"
)

// HTTP хендлер для приложения
type Handler struct {
	userService *service.UserService
	logger      *logger.Logger
}

// констуктор для HTTP хендлера
func NewHandler(u *service.UserService, l *logger.Logger) *Handler {
	return &Handler{
		userService: u,
		logger:      l,
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
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if req.Login == "" || req.Password == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		token, err := h.userService.Register(req.Login, req.Password)
		if err != nil {
			switch err {
			case repository.ErrUserExists:
				http.Error(w,
					http.StatusText(http.StatusConflict),
					http.StatusConflict,
				)
			default:
				h.logger.Error("failed to register user", "error", err)

				http.Error(w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
			}

			return
		}

		w.Header().Set("Authorization", token)
		w.WriteHeader(http.StatusOK)
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
			switch err {
			case service.ErrInvalidCredentials:
				http.Error(w,
					http.StatusText(http.StatusUnauthorized),
					http.StatusUnauthorized,
				)
			default:
				h.logger.Error("failed to login user", "error", err)

				http.Error(w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
			}

			return
		}

		w.Header().Set("Authorization", token)
		w.WriteHeader(http.StatusOK)
	}
}
