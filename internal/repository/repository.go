package repository

import (
	"errors"

	"github.com/onbehalfofhim/gofermart/internal/models"
)

// интерфейс хранилища пользователей приложения
type UserRepo interface {
	Create(login, passwordHash string) (*models.User, error)
	GetByLogin(login string) (*models.User, error)
}

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")
