package services

import (
	"encoding/json"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"
	"time"
)

var loc, _ = time.LoadLocation("Asia/Ho_Chi_Minh")

type OrderService interface {
	GetAllOrder(query dto.OrderQuery) ([]dto.OrderReponse, int64, error)
	GetOrder(id int64) (*dto.OrderReponse, error)
	IsDriverAssignedToOrder(orderID int64, driverID int64) (bool, error)
	CreateOrder(req dto.OrderRequest, updatedBy string) (*dto.CreateOrderResponse, error)
	UpdateOrderStatus(id int64, status, updatedBy string, driverID *int64) (*models.Order, error)
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
		userInfo := &models.UserInfo{}
		if len(order.UserInfo) > 0 {
			json.Unmarshal(order.UserInfo, userInfo)
		}

		response = append(response, dto.OrderReponse{
			ID:              order.ID,
			TotalAmount:     order.TotalAmount,
			Username:        userInfo.Username,
			UserPhone:       userInfo.UserPhone,
			ShippingAddress: userInfo.ShippingAddress,
			Status:          order.CurrentStatus,
			Ordered_at:      order.CreatedAt.In(loc),
		})
	}
	return response, total, nil
}

func (s *orderService) GetOrder(id int64) (*dto.OrderReponse, error) {
	order, err := s.orderRepo.GetOrderDetail(id)
	if err != nil {

		return nil, err
	}

	userInfo := &models.UserInfo{}

	if len(order.UserInfo) > 0 {
		json.Unmarshal(order.UserInfo, userInfo)
	}

	response := dto.OrderReponse{
		ID:              order.ID,
		TotalAmount:     order.TotalAmount,
		Username:        userInfo.Username,
		UserPhone:       userInfo.UserPhone,
		ShippingAddress: userInfo.ShippingAddress,
		Status:          order.CurrentStatus,
		Ordered_at:      order.CreatedAt.In(loc),
	}

	return &response, nil
}

func (s *orderService) IsDriverAssignedToOrder(orderID int64, driverID int64) (bool, error) {
	return s.orderRepo.IsDriverAssignedToOrder(orderID, driverID)
}

func (s *orderService) CreateOrder(req dto.OrderRequest, updatedBy string) (*dto.CreateOrderResponse, error) {
	userInfo := models.UserInfo{
		Username:        req.Username,
		UserPhone:       req.UserPhone,
		ShippingAddress: req.ShippingAddress,
	}

	userInfoJSON, _ := json.Marshal(userInfo)

	order := models.Order{
		UserInfo:      userInfoJSON,
		TotalAmount:   req.TotalAmount,
		CurrentStatus: models.ORDER_STATUS_CREATED,
	}

	saved, err := s.orderRepo.CreateOrder(order, updatedBy)
	if err != nil {
		return nil, err
	}

	return &dto.CreateOrderResponse{
		ID:              saved.ID,
		Status:          saved.CurrentStatus,
		TotalAmount:     saved.TotalAmount,
		ShippingAddress: req.ShippingAddress,
		CreatedAt:       saved.CreatedAt,
		UpdatedAt:       saved.UpdatedAt,
	}, nil
}

func (s *orderService) UpdateOrderStatus(id int64, status, updatedBy string, driverID *int64) (*models.Order, error) {
	newOrder, err := s.orderRepo.UpdateOrderStatus(id, status, updatedBy, driverID)
	if err != nil {
		return nil, err
	}

	return newOrder, nil
}
