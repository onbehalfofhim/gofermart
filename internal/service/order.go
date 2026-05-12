package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/onbehalfofhim/gofermart/internal/models"
	"github.com/onbehalfofhim/gofermart/internal/repository"
)

var ErrOrderBelongsToOtherUser = errors.New("order belongs to another user")
var ErrUnknownOrderDtatus = errors.New("order status is not matched")

// сервис работы с заказом
type OrderService struct {
	repo repository.OrderRepo
}

// конструктор для создания сервиса
func NewOrderService(r repository.OrderRepo) *OrderService {
	return &OrderService{
		repo: r,
	}
}

// создание заказа с потенциальным начислением баллов лояльности
func (s *OrderService) Create(ctx context.Context, number string, userId uuid.UUID) error {
	// Пытаемся создать заказ
	order, err := s.repo.Create(ctx, number, userId)

	if err != nil {
		switch err {
		case repository.ErrOrderExists:
			if order.UserID != userId {
				return ErrOrderBelongsToOtherUser
			}
			return err
		default:
			return err
		}
	}

	return nil
}

// получение списка заказов на обработку
func (s *OrderService) GetOrdersForProcessing(ctx context.Context) ([]models.Order, error) {
	return s.repo.GetOrdersForProcessing(ctx)
}

// обновление статуса заказа в системе
func (s *OrderService) UpdateStatus(ctx context.Context, orderNumber string, status string, accrual *float64) error {
	return s.repo.UpdateStatus(ctx, orderNumber, status, accrual)
}

// получение списка закзаов пользователя
func (s *OrderService) GetOrdersByUserId(ctx context.Context, userId uuid.UUID) ([]models.OrderResponse, error) {
	orders, err := s.repo.GetOrdersByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	// Преобразуем в response формат
	responses := make([]models.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = models.OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			UploadedAt: order.UploadedAt.Format("2006-01-02T15:04:05Z07:00"), // RFC3339
		}

		// Добавляем accrual только если он больше 0
		if order.Accrual > 0 {
			responses[i].Accrual = &order.Accrual
		}
	}

	return responses, nil
}
