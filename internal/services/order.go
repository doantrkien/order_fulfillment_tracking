package services

import (
	"encoding/json"
	"errors"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"
	"main/pkg/utils/constant"

	"gorm.io/gorm"
)

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
			Ordered_at:      order.CreatedAt,
		})
	}
	return response, total, nil
}

func (s *orderService) GetOrder(id int64) (*dto.OrderReponse, error) {
	order, err := s.orderRepo.GetOrderDetail(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constant.ERR_NOT_FOUND
		}

		return nil, err
	}

	if order == nil {
		return nil, constant.ERR_NOT_FOUND
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
		Ordered_at:      order.CreatedAt,
	}

	return &response, nil
}

func (s *orderService) CreateOrder(req dto.OrderRequest) (*models.Order, error) {
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

	return s.orderRepo.CreateOrder(order)
}

func (s *orderService) UpdateOrderStatus(id int64, status string) (*models.Order, error) {
	order, err := s.GetOrder(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constant.ERR_NOT_FOUND
		}
		return nil, err
	}

	if order == nil {
		return nil, constant.ERR_NOT_FOUND
	}

	if !models.IsValidTransition(order.Status, models.OrderStatus(status)) {
		return nil, constant.ORDER_STATUS_TRANSITION_INVALID
	}

	newOrder, err := s.orderRepo.UpdateOrderStatus(id, status)
	if err != nil {
		return nil, err
	}

	return newOrder, nil
}
