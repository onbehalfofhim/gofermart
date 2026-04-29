package inmemory

import (
	"github.com/google/uuid"
	"github.com/onbehalfofhim/gofermart/internal/models"
	"github.com/onbehalfofhim/gofermart/internal/repository"
)

type UsersMemStorage struct {
	users map[string]models.User
}

func NewUsersMemStorage() *UsersMemStorage {
	return &UsersMemStorage{
		users: make(map[string]models.User, 0),
	}
}

func (r *UsersMemStorage) Create(login, passwordHash string) (*models.User, error) {
	if _, ok := r.users[login]; ok {
		return nil, repository.ErrUserExists
	}

	user := models.User{
		ID:           uuid.New(),
		Login:        login,
		PasswordHash: passwordHash,
	}
	r.users[login] = user
	return &user, nil

}

func (r *UsersMemStorage) GetByLogin(login string) (*models.User, error) {
	user, ok := r.users[login]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return &user, nil

}
