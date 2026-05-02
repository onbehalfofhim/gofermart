package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/onbehalfofhim/gofermart/internal/auth"
	"github.com/onbehalfofhim/gofermart/internal/logger"
	"github.com/onbehalfofhim/gofermart/internal/middleware"
	"github.com/onbehalfofhim/gofermart/internal/models"
	"github.com/onbehalfofhim/gofermart/internal/repository"
	"github.com/onbehalfofhim/gofermart/internal/service"
)

// HTTP хендлер для приложения
type Handler struct {
	userService  *service.UserService
	orderService *service.OrderService
	logger       *logger.Logger
	jwt          *auth.JWT
}

// констуктор для HTTP хендлера
func NewHandler(u *service.UserService, o *service.OrderService, l *logger.Logger, j *auth.JWT) *Handler {
	return &Handler{
		userService:  u,
		orderService: o,
		logger:       l,
		jwt:          j,
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

		user, err := h.userService.Register(req.Login, req.Password)
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

		// Генерируем JWT токена
		token, err := h.jwt.GenerateToken(user.ID.String())
		if err != nil {
			h.logger.Error("failed to generate token", "error", err)

			http.Error(w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
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

		user, err := h.userService.Login(req.Login, req.Password)
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

		// Генерируем JWT токена
		token, err := h.jwt.GenerateToken(user.ID.String())
		if err != nil {
			h.logger.Error("failed to generate token", "error", err)

			http.Error(w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
		}

		w.Header().Set("Authorization", token)
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) CreateOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		userId, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		orderNumber := strings.TrimSpace(string(body))
		if orderNumber == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// TBD: проверить номер заказа по Алгоритму Луна

		err = h.orderService.Create(orderNumber, userId)
		if err != nil {
			switch err {
			case repository.ErrOrderExists:
				http.Error(w,
					http.StatusText(http.StatusOK),
					http.StatusOK,
				)
			case service.ErrOrderBelongsToOtherUser:
				http.Error(w,
					http.StatusText(http.StatusConflict),
					http.StatusConflict,
				)
			default:
				h.logger.Error("failed to create order", "error", err)

				http.Error(w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
			}
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
