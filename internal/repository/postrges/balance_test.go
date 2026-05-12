package postrges

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/onbehalfofhim/gofermart/internal/models"
	"github.com/onbehalfofhim/gofermart/internal/repository"
)

func TestBalanceRepository_CreateAccrual(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewBalanceRepository(db)

	userId := uuid.New()

	mock.ExpectBegin()

	mock.ExpectExec(`UPDATE users`).
		WithArgs(100.0, userId).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	mock.ExpectExec(`INSERT INTO balance_operations`).
		WithArgs(
			sqlmock.AnyArg(),
			userId,
			"12345678903",
			models.OperationAccrual,
			100.0,
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	mock.ExpectCommit()

	err = repo.CreateAccrual(context.Background(), "12345678903", userId, 100)

	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_CreateWithdraw(t *testing.T) {
	tests := []struct {
		name         string
		rowsAffected int64
		expectedErr  error
	}{
		{
			name:         "success",
			rowsAffected: 1,
			expectedErr:  nil,
		},
		{
			name:         "insufficient funds",
			rowsAffected: 0,
			expectedErr:  repository.ErrInsufficientFunds,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewBalanceRepository(db)

			userId := uuid.New()

			mock.ExpectBegin()

			mock.ExpectExec(`UPDATE users`).
				WithArgs(50.0, userId).
				WillReturnResult(
					sqlmock.NewResult(
						0,
						tt.rowsAffected,
					),
				)

			if tt.rowsAffected > 0 {
				mock.ExpectExec(
					`INSERT INTO balance_operations`,
				).
					WithArgs(
						sqlmock.AnyArg(),
						userId,
						"12345678903",
						models.OperationWithdraw,
						50.0,
						sqlmock.AnyArg(),
					).
					WillReturnResult(
						sqlmock.NewResult(0, 1),
					)

				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}

			err = repo.CreateWithdraw(context.Background(), "12345678903", userId, 50)

			assert.ErrorIs(t, err, tt.expectedErr)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestBalanceRepository_GetBalanceWithWithdrawn(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewBalanceRepository(db)

	userID := uuid.New()

	rows := sqlmock.NewRows(
		[]string{
			"balance",
			"withdrawn",
		},
	).AddRow(500.0, 200.0)

	mock.ExpectQuery(`SELECT`).
		WithArgs(userID).
		WillReturnRows(rows)

	balance, withdrawn, err := repo.GetBalanceWithWithdrawn(context.Background(), userID)

	assert.NoError(t, err)

	assert.Equal(t, 500.0, balance)
	assert.Equal(t, 200.0, withdrawn)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBalanceRepository_GetWithdrawalsByUserId(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewBalanceRepository(db)

	userId := uuid.New()

	now := time.Now()

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"order_number",
			"user_id",
			"operation_type",
			"amount",
			"processed_at",
			"created_at",
		},
	).AddRow(
		uuid.New(),
		"12345678903",
		userId,
		models.OperationWithdraw,
		300.0,
		now,
		now,
	)

	mock.ExpectQuery(`SELECT id, order_number`).
		WithArgs(userId).
		WillReturnRows(rows)

	operations, err := repo.GetWithdrawalsByUserId(context.Background(), userId)
	assert.NoError(t, err)

	assert.Len(t, operations, 1)

	assert.Equal(t, "12345678903", operations[0].OrderNumber)

	assert.Equal(t, models.OperationWithdraw, operations[0].OperationType)

	assert.Equal(t, 300.0, operations[0].Amount)

	assert.NoError(t, mock.ExpectationsWereMet())
}
