package services

import (
	"main/internal/dto"
	"main/internal/models"
)

type OrderService interface {
	GetAllOrder() ([]dto.OrderReponse, error)
	GetOrder(id int) (*dto.OrderReponse, error)
}

type orderService struct {
	orderRepo models.OrderRepository
}

func NewOrderService(orderRepo models.OrderRepository) OrderService {
	return &orderService{
		orderRepo: orderRepo,
	}
}

func (s *orderService) GetAllOrder() ([]dto.OrderReponse, error) {
	orders, err := s.orderRepo.GetAllOrder()
	if err != nil {
		return nil, err
	}

	var response []dto.OrderReponse

	for _, order := range orders {
		response = append(response, dto.OrderReponse{
			TotalAmount:  order.TotalAmount,
			ShippingAddr: order.ShippingAddr,
			Status:       order.Status,
		})
	}

	return response, nil
}

func (s *orderService) GetOrder(id int) (*dto.OrderReponse, error) {
	order, err := s.orderRepo.GetOrderDetail(id)
	if err != nil {
		return nil, err
	}

	response := dto.OrderReponse{
		TotalAmount:  order.TotalAmount,
		ShippingAddr: order.ShippingAddr,
		Status:       order.Status,
	}

	return &response, nil
}
