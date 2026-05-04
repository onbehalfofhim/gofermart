package postrges

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/onbehalfofhim/gofermart/internal/models"
	"github.com/onbehalfofhim/gofermart/internal/repository"
)

// представляет репозиторий для работы с заказами
type BalanceRepository struct {
	db *sql.DB
}

// создает новый репозиторий заказов
func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) CreateAccrual(ctx context.Context, orderNumber string, userId uuid.UUID, amount float64) error {
	// начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// увеличиваем баланс
	queryUsers := `UPDATE users 
		SET balance = balance + $1 
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, queryUsers, amount, userId)
	if err != nil {
		return err
	}

	// записываем операцию
	queryOperation := `INSERT INTO balance_operations 
        (id, user_id, order_number, operation_type, amount, processed_at)
        VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(
		ctx,
		queryOperation,
		uuid.New(),
		userId,
		orderNumber,
		models.OperationAccrual,
		amount,
		time.Now(),
	)
	if err != nil {
		return err
	}

	// завершаем транзакцию
	return tx.Commit()
}

func (r *BalanceRepository) CreateWithdraw(ctx context.Context, orderNumber string, userId uuid.UUID, amount float64) error {
	// начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// атомарное списание
	queryUsers := `UPDATE users
         SET balance = balance - $1
         WHERE id = $2 AND balance >= $1
	`
	res, err := tx.ExecContext(ctx, queryUsers, amount, userId)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return repository.ErrInsufficientFunds
	}

	// записываем операцию
	queryOperation := `INSERT INTO balance_operations 
        (id, user_id, order_number, operation_type, amount, processed_at)
        VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(
		ctx,
		queryOperation,
		uuid.New(),
		userId,
		orderNumber,
		models.OperationWithdraw,
		amount,
		time.Now(),
	)
	if err != nil {
		return err
	}

	// завершаем транзакцию
	return tx.Commit()
}

func (r *BalanceRepository) GetBalanceWithWithdrawn(ctx context.Context, userId uuid.UUID) (float64, float64, error) {
	query := `SELECT 
            u.balance,
            COALESCE(SUM(
                CASE 
                    WHEN bo.operation_type = 'WITHDRAW' THEN bo.amount 
                    ELSE 0 
                END
            ), 0)
        FROM users u
        LEFT JOIN balance_operations bo ON bo.user_id = u.id
        WHERE u.id = $1
        GROUP BY u.balance
    `
	row := r.db.QueryRowContext(ctx, query, userId)

	var balance float64
	var withdrawn float64
	err := row.Scan(&balance, &withdrawn)

	if err != nil {
		return 0, 0, err
	}

	return balance, withdrawn, nil
}
