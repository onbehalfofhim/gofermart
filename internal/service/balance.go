package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/onbehalfofhim/gofermart/internal/models"
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

func (s *BalanceService) GetBalanceWithWithdrawn(ctx context.Context, userId uuid.UUID) (models.BalanceResponse, error) {
	balance, withdrawn, err := s.repo.GetBalanceWithWithdrawn(ctx, userId)
	if err != nil {
		return models.BalanceResponse{}, err
	}

	response := models.BalanceResponse{
		Current:   balance,
		Withdrawn: withdrawn,
	}

	return response, nil
}

func (s *BalanceService) GetWithdrawalsByUserId(ctx context.Context, userId uuid.UUID) ([]models.WithdrawalResponse, error) {
	operations, err := s.repo.GetWithdrawalsByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	// Преобразуем в response формат
	responses := make([]models.WithdrawalResponse, len(operations))
	for i, withdraw := range operations {
		responses[i] = models.WithdrawalResponse{
			Order:       withdraw.OrderNumber,
			Sum:         withdraw.Amount,
			ProcessedAt: withdraw.ProcessedAt.Format("2006-01-02T15:04:05Z07:00"), // RFC3339
		}
	}

	return responses, nil
}
