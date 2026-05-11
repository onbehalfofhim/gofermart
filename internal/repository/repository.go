package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/onbehalfofhim/gofermart/internal/models"
)

// интерфейс хранилища пользователей приложения
type UserRepo interface {
	Create(ctx context.Context, login, passwordHash string) (*models.User, error)
	GetByLogin(ctx context.Context, login string) (*models.User, error)
}

// интерфейс хранилища пользователей приложения
type OrderRepo interface {
	Create(ctx context.Context, number string, userId uuid.UUID) (*models.Order, error)
	UpdateStatus(ctx context.Context, number string, status string, accrual *float64) error

	GetOrdersForProcessing(ctx context.Context) ([]models.Order, error)
}

// интерфейс хранилища пользователей приложения
type BalanceRepo interface {
	CreateAccrual(ctx context.Context, orderNumber string, userId uuid.UUID, amount float64) error
	CreateWithdraw(ctx context.Context, orderNumber string, userId uuid.UUID, amount float64) error
	GetBalanceWithWithdrawn(ctx context.Context, userId uuid.UUID) (float64, float64, error)
}

var (
	// ошибки репозитория с пользоватлями
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")

	// ошибки репозитория с заказами
	ErrOrderExists   = errors.New("order number already exists")
	ErrOrderNotFound = errors.New("order not found")

	// ошибки репозитория с операциями баланса
	ErrInsufficientFunds = errors.New("insufficient funds")
)
