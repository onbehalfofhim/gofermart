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

// представляет репозиторий для работы с заказами
type OrdersRepository struct {
	db *sql.DB
}

// создает новый репозиторий заказов
func NewOrdersRepository(db *sql.DB) *OrdersRepository {
	return &OrdersRepository{db: db}
}

func (r *OrdersRepository) Create(ctx context.Context, number string, userId uuid.UUID) (*models.Order, error) {
	query := `INSERT INTO orders (id, number, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, number, user_id, status, accrual, uploaded_at
	`

	row := r.db.QueryRowContext(ctx, query, uuid.New(), number, userId)

	var order models.Order
	err := row.Scan(
		&order.ID,
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		// проверка, что полученная ошибка - ошибка уникальности номера
		// (=заказ с таким номером уже создавался)
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				// получить существующий заказ по номеру
				existingOrder, error := r.getOrderByNumber(ctx, number)
				if error != nil {
					return nil, error
				}
				return existingOrder, repository.ErrOrderExists
			}
		}
		return nil, err
	}

	return &order, nil
}

func (r *OrdersRepository) getOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	query := `SELECT id, number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE number = $1
	`

	var order models.Order
	row := r.db.QueryRowContext(ctx, query, number)

	err := row.Scan(
		&order.ID,
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrOrderNotFound
		}
		return nil, err
	}

	return &order, nil
}
