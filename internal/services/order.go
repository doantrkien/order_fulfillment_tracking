package services

import (
	"main/internal/models"
	"main/internal/repositories"
)

type OrderService struct {
	orderRepo *repositories.OrderRepository
}

func NewOrderService(orderRepo *repositories.OrderRepository) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
	}
}

func (s *OrderService) GetAllOrder() ([]models.Order, error) {
	return s.orderRepo.GetAllOrder()
}
