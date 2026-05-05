package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/onbehalfofhim/gofermart/internal/repository"
)

var ErrOrderBelongsToOtherUser = errors.New("order belongs to another user")

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
