package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/onbehalfofhim/gofermart/internal/models"
)

// интерфейс хранилища пользователей приложения
type UserRepo interface {
	Create(login, passwordHash string) (*models.User, error)
	GetByLogin(login string) (*models.User, error)
}

// интерфейс хранилища пользователей приложения
type OrderRepo interface {
	Create(number string, userId uuid.UUID) (*models.Order, error)
}

var (
	// ошибки репозитория с пользоватлями
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")

	// ошибки репозитория с заказами
	ErrOrderExists   = errors.New("order number already exists")
	ErrOrderNotFound = errors.New("order not found")
)
