package services

import (
	"encoding/json"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"
)

type OrderService interface {
	GetAllOrder(query dto.OrderQuery) ([]dto.OrderReponse, int64, error)
	GetOrder(id int) (*dto.OrderReponse, error)
	CreateOrder(dto.OrderRequest) (*models.Order, error)
	UpdateOrderStatus(id int64, status string) (*models.Order, error)
}

type orderService struct {
	orderRepo repositories.OrderRepository
}

func NewOrderService(orderRepo repositories.OrderRepository) OrderService {
	return &orderService{
		orderRepo: orderRepo,
	}
}

func (s *orderService) GetAllOrder(query dto.OrderQuery) ([]dto.OrderReponse, int64, error) {

	orders, total, err := s.orderRepo.GetAllOrder(query)
	if err != nil {
		return nil, 0, err
	}

	var response []dto.OrderReponse

	for _, order := range orders {

		userInfo := make(map[string]interface{})

		err := json.Unmarshal(order.UserInfo, &userInfo)
		if err != nil {
			return nil, 0, err
		}

		response = append(response, dto.OrderReponse{
			CustomerName:  userInfo["customer_name"].(string),
			CustomerPhone: userInfo["customer_phone"].(string),
			TotalAmount:   order.TotalAmount,
			ShippingAddr:  userInfo["shipping_addr"].(string),
			Status:        order.CurrentStatus,
			Ordered_at:    order.CreatedAt,
		})
	}
	return response, total, nil
}

func (s *orderService) GetOrder(id int) (*dto.OrderReponse, error) {
	order, err := s.orderRepo.GetOrderDetail(id)
	if err != nil {
		return nil, err
	}

	userInfo := make(map[string]interface{})

	err = json.Unmarshal(order.UserInfo, &userInfo)
	if err != nil {
		return nil, err
	}

	response := dto.OrderReponse{
		CustomerName:  userInfo["customer_name"].(string),
		CustomerPhone: userInfo["customer_phone"].(string),
		TotalAmount:   order.TotalAmount,
		ShippingAddr:  userInfo["shipping_addr"].(string),
		Status:        order.CurrentStatus,
		Ordered_at:    order.CreatedAt,
	}

	return &response, nil
}

func (s *orderService) CreateOrder(req dto.OrderRequest) (*models.Order, error) {
	userInfo := map[string]interface{}{
		"customer_name":  req.CustomerName,
		"customer_phone": req.CustomerPhone,
		"shipping_addr":  req.ShippingAddr,
	}

	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return nil, err
	}

	order := models.Order{
		UserInfo:      userInfoJSON,
		TotalAmount:   req.TotalAmount,
		CurrentStatus: models.ORDER_STATUS_CREATED,
	}

	return s.orderRepo.CreateOrder(order)
}

func (s *orderService) UpdateOrderStatus(id int64, status string) (*models.Order, error) {
	order, err := s.orderRepo.UpdateOrderStatus(id, status)
	if err != nil {
		return nil, err
	}
	return order, nil
}
