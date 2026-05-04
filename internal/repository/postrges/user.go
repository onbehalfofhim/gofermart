package postrges

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/onbehalfofhim/gofermart/internal/models"
	"github.com/onbehalfofhim/gofermart/internal/repository"
)

// представляет репозиторий для работы с пользователями
type UsersRepository struct {
	db *sql.DB
}

// создает новый репозиторий пользователей
func NewUsersRepository(db *sql.DB) *UsersRepository {
	return &UsersRepository{db: db}
}

// создание пользователя в БД
func (r *UsersRepository) Create(ctx context.Context, login, passwordHash string) (*models.User, error) {
	query := `INSERT INTO users (id, login, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, login, password_hash, created_at
	`

	row := r.db.QueryRowContext(ctx, query, uuid.New(), login, passwordHash)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		// проверка, что полученная ошибка - ошибка уникальности логина
		// (=пользователь с таким логикном уже есть)
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, repository.ErrUserExists
			}
		}
		return nil, err
	}

	return &user, nil
}

// поиск пользователя по логину
func (r *UsersRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1
	`

	row := r.db.QueryRowContext(ctx, query, login)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
