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
	"github.com/onbehalfofhim/gofermart/pkg/utils"
)

// HTTP хендлер для приложения
type Handler struct {
	userService    *service.UserService
	orderService   *service.OrderService
	balanceService *service.BalanceService
	logger         *logger.Logger
	jwt            *auth.JWT
}

// констуктор для HTTP хендлера
func NewHandler(u *service.UserService, o *service.OrderService, b *service.BalanceService, l *logger.Logger, j *auth.JWT) *Handler {
	return &Handler{
		userService:    u,
		orderService:   o,
		balanceService: b,
		logger:         l,
		jwt:            j,
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

		user, err := h.userService.Register(r.Context(), req.Login, req.Password)
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

		user, err := h.userService.Login(r.Context(), req.Login, req.Password)
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

// обработчик регистрации заказа в системе лояльности
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

		if !utils.ValidateLuhn(orderNumber) {
			http.Error(w,
				http.StatusText(http.StatusUnprocessableEntity),
				http.StatusUnprocessableEntity,
			)
			return
		}

		err = h.orderService.Create(r.Context(), orderNumber, userId)
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

// обработчик регистрации списания баллов
func (h *Handler) CreateWithdraw() http.HandlerFunc {
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

		var req models.WithdrawRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if !utils.ValidateLuhn(req.Order) {
			http.Error(w,
				http.StatusText(http.StatusUnprocessableEntity),
				http.StatusUnprocessableEntity,
			)
			return
		}

		err = h.balanceService.CreateWithdraw(r.Context(), req.Order, userId, req.Sum)
		if err != nil {
			switch err {
			case repository.ErrInsufficientFunds:
				http.Error(w,
					http.StatusText(http.StatusPaymentRequired),
					http.StatusPaymentRequired,
				)
			default:
				h.logger.Error("failed to create withdraw", "error", err)

				http.Error(w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) GetOrders() http.HandlerFunc {
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

		w.Header().Set("Content-Type", "application/json")

		resp, err := h.orderService.GetOrdersByUserId(r.Context(), userId)
		if err != nil {
			h.logger.Error("failed to get orders", "error", err)

			http.Error(w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}

		if len(resp) == 0 {
			http.Error(w,
				http.StatusText(http.StatusNoContent),
				http.StatusNoContent,
			)
		}

		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			http.Error(w, "cannot encode response body", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}
