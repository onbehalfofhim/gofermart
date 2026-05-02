package service

import (
	"errors"
	"fmt"

	"github.com/onbehalfofhim/gofermart/internal/auth"
	"github.com/onbehalfofhim/gofermart/internal/models"
	"github.com/onbehalfofhim/gofermart/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// сервис работы с пользователем
type UserService struct {
	repo repository.UserRepo
}

// конструктор для создания сервиса
func NewUserService(r repository.UserRepo) *UserService {
	return &UserService{
		repo: r,
	}
}

// Регистрирация нового пользователя
func (s *UserService) Register(login, password string) (*models.User, error) {
	// Хешируем пароль
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	user, err := s.repo.Create(login, hash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// аутентификация пользователя
func (s *UserService) Login(login, password string) (*models.User, error) {
	// Получаем пользователя по логину
	user, err := s.repo.GetByLogin(login)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Проверяем пароль
	err = auth.CheckPassword(user.PasswordHash, password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
