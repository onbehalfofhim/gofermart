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

// записть заказа в БД
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

// получение информации о заказе по его номеру
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

// получение заказов, находящихся не в коненом статусе статусной модели
func (r *OrdersRepository) GetOrdersForProcessing(ctx context.Context) ([]models.Order, error) {
	query := `SELECT id, number, user_id, status, accrual, uploaded_at, updated_at
		FROM orders
		WHERE status IN ($1, $2)
	`

	rows, err := r.db.QueryContext(ctx, query, models.OrderStatusNew, models.OrderStatusProcessing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// маппинг внешних статусов на статусную модель системы
func (r *OrdersRepository) MapStatus(status string) models.OrderStatus {
	switch status {
	case "REGISTERED":
		return models.OrderStatusProcessing
	case "PROCESSING":
		return models.OrderStatusProcessing
	case "INVALID":
		return models.OrderStatusInvalid
	case "PROCESSED":
		return models.OrderStatusProcessed
	default:
		return models.OrderStatusProcessing
	}
}

// обновление статуса заказа
func (r *OrdersRepository) UpdateStatus(ctx context.Context, number string, status string, accrual *float64) error {
	var (
		query string
		args  []any
	)

	orderStatus := r.MapStatus(status)

	if accrual != nil {
		query = `UPDATE orders
			SET
				status = $1,
				accrual = $2,
				updated_at = NOW()
			WHERE number = $3
		`
		args = []any{orderStatus, *accrual, number}
	} else {
		query = `UPDATE orders
			SET
				status = $1,
				updated_at = NOW()
			WHERE number = $2
		`
		args = []any{orderStatus, number}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrOrderNotFound
	}

	return nil
}

// получение списка заказов пользователя
func (r *OrdersRepository) GetOrdersByUserId(ctx context.Context, userId uuid.UUID) ([]models.Order, error) {
	query := `SELECT id, number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
