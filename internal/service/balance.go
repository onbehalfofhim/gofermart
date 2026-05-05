package service

import (
	"context"

	"github.com/google/uuid"
	// "github.com/onbehalfofhim/gofermart/internal/models"
	"github.com/onbehalfofhim/gofermart/internal/repository"
)

// сервис работы с заказом
type BalanceService struct {
	repo repository.BalanceRepo
}

// конструктор для создания сервиса
func NewBalanceService(r repository.BalanceRepo) *BalanceService {
	return &BalanceService{
		repo: r,
	}
}

// создаем операцию с балансом (начисление)
func (s *BalanceService) CreateAccrual(ctx context.Context, orderNumber string, userId uuid.UUID, accrual float64) error {
	return s.repo.CreateAccrual(ctx, orderNumber, userId, accrual)
}

// создаем операцию с балансом (списание)
func (s *BalanceService) CreateWithdraw(ctx context.Context, orderNumber string, userId uuid.UUID, sum float64) error {
	return s.repo.CreateWithdraw(ctx, orderNumber, userId, sum)
}
