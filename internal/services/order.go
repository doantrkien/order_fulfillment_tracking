package services

import (
	"main/errs"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"
	"time"
)

var loc, _ = time.LoadLocation("Asia/Ho_Chi_Minh")

type OrderService interface {
	GetAllOrder(query dto.OrderQuery) ([]dto.OrderReponse, int64, error)
	GetOrder(id int64) (*dto.OrderReponse, error)
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

func orderToResponse(order models.Order) dto.OrderReponse {
	resp := dto.OrderReponse{
		ID:          order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      order.CurrentStatus,
		Ordered_at:  order.CreatedAt.In(loc),
	}
	if order.User != nil {
		resp.Username = order.User.Username
		resp.UserPhone = order.User.Phone
		resp.ShippingAddress = order.User.Address
	}
	return resp
}

func (s *orderService) GetAllOrder(query dto.OrderQuery) ([]dto.OrderReponse, int64, error) {
	orders, total, err := s.orderRepo.GetAllOrder(query)
	if err != nil {
		return nil, 0, err
	}

	var response []dto.OrderReponse
	for _, order := range orders {
		response = append(response, orderToResponse(order))
	}
	return response, total, nil
}

func (s *orderService) GetOrder(id int64) (*dto.OrderReponse, error) {
	order, err := s.orderRepo.GetOrderDetail(id)
	if err != nil {
		return nil, err
	}
	resp := orderToResponse(*order)
	return &resp, nil
}

func (s *orderService) CreateOrder(req dto.OrderRequest) (*models.Order, error) {
	order := models.Order{
		UserID:        req.UserID,
		TotalAmount:   req.TotalAmount,
		CurrentStatus: models.ORDER_STATUS_CREATED,
	}
	return s.orderRepo.CreateOrder(order)
}

func (s *orderService) UpdateOrderStatus(id int64, status string) (*models.Order, error) {
	order, err := s.GetOrder(id)
	if err != nil {
		return nil, err
	}

	if !models.IsValidTransition(order.Status, models.OrderStatus(status)) {
		return nil, errs.ERR_INVALID_STATUS
	}

	return s.orderRepo.UpdateOrderStatus(id, status)
}
