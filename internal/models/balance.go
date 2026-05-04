package models

import (
	"time"

	"github.com/google/uuid"
)

// OperationType представляет тип операции с баллами
type OperationType string

const (
	OperationAccrual  OperationType = "ACCRUAL"
	OperationWithdraw OperationType = "WITHDRAW"
)

// модель для хранения операций с баллами лояльности
type BalanceOperation struct {
	ID            uuid.UUID     `json:"id"`
	OrderNumber   string        `json:"order_number"`
	UserID        uuid.UUID     `json:"user_id"`
	OperationType OperationType `json:"operation_type"`
	Amount        float64       `json:"amount"`
	ProcessedAt   time.Time     `json:"processed_at"`
	CreatedAt     time.Time     `json:"created_at"`
}

// модель ответа с информацией о балансе пользователя
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// модель запроса на списание средств
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// модель ответа с информацией о списании
type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
