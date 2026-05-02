package service

import (
	"errors"
	"fmt"

	"github.com/onbehalfofhim/gofermart/internal/auth"
	"github.com/onbehalfofhim/gofermart/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// сервис работы с пользователем
type UserService struct {
	repo       repository.UserRepo
	jwtManager *auth.JWT
}

// конструктор для создания сервиса
func NewUserService(r repository.UserRepo, j *auth.JWT) *UserService {
	return &UserService{
		repo:       r,
		jwtManager: j,
	}
}

// Регистрирация нового пользователя
func (s *UserService) Register(login, password string) (string, error) {
	// Хешируем пароль
	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	user, err := s.repo.Create(login, hash)
	if err != nil {
		return "", err
	}

	// Генерируем JWT токена
	token, err := s.jwtManager.GenerateToken(user.ID.String())
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// аутентификация пользователя
func (s *UserService) Login(login, password string) (string, error) {
	// Получаем пользователя по логину
	user, err := s.repo.GetByLogin(login)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	// Проверяем пароль
	err = auth.CheckPassword(user.PasswordHash, password)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// Генерируем JWT токена
	token, err := s.jwtManager.GenerateToken(user.ID.String())
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}
