package service

import (
	"errors"
	"fmt"

	"github.com/onbehalfofhim/gofermart/internal/auth"
	// "github.com/onbehalfofhim/gofermart/internal/models"
	"github.com/onbehalfofhim/gofermart/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type UserService struct {
	repo repository.UserRepo
}

func NewUserService(r repository.UserRepo) *UserService {
	return &UserService{repo: r}
}

// Register регистрирует нового пользователя
func (s *UserService) Register(login, password string) error {
	// Хешируем пароль
	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	_, err = s.repo.Create(login, hash)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) Login(login, password string) error {
	// Получаем пользователя по логину
	user, err := s.repo.GetByLogin(login)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return ErrInvalidCredentials
		}
		return err
	}

	// Проверяем пароль
	err = auth.CheckPassword(user.PasswordHash, password)
	if err != nil {
		return ErrInvalidCredentials
	}

	return nil
}
