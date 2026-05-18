package services_test

import (
	"encoding/json"
	"main/internal/dto"
	"main/internal/mocks"
	"main/internal/models"
	"main/internal/services"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupServiceTest(t *testing.T) (*mocks.OrderRepository, services.OrderService) {
	mockRepo := mocks.NewOrderRepository(t)
	service := services.NewOrderService(mockRepo)
	return mockRepo, service
}

func TestOrderService_CreateOrder(t *testing.T) {
	mockRepo, service := setupServiceTest(t)

	t.Run("Success", func(t *testing.T) {
		req := dto.OrderRequest{
			TotalAmount:     2000,
			Username:        "testuser",
			UserPhone:       "0987654321",
			ShippingAddress: "Test Address",
		}

		userInfo := models.UserInfo{
			Username:        req.Username,
			UserPhone:       req.UserPhone,
			ShippingAddress: req.ShippingAddress,
		}
		userInfoJSON, _ := json.Marshal(userInfo)

		expectedOrder := models.Order{
			UserInfo:      userInfoJSON,
			TotalAmount:   req.TotalAmount,
			CurrentStatus: models.ORDER_STATUS_CREATED,
		}

		mockRepo.On("CreateOrder", expectedOrder).Return(&models.Order{
			ID:            1,
			TotalAmount:   2000,
			CurrentStatus: models.ORDER_STATUS_CREATED,
			UserInfo:      userInfoJSON,
		}, nil).Once()

		createdOrder, err := service.CreateOrder(req)

		assert.NoError(t, err)
		assert.NotNil(t, createdOrder)
		assert.Equal(t, int64(1), createdOrder.ID)
		assert.Equal(t, req.TotalAmount, createdOrder.TotalAmount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		req := dto.OrderRequest{TotalAmount: 2000}
		mockRepo.On("CreateOrder", mock.Anything).Return(nil, assert.AnError).Once()

		createdOrder, err := service.CreateOrder(req)

		assert.Error(t, err)
		assert.Nil(t, createdOrder)
		mockRepo.AssertExpectations(t)
	})
}

func TestOrderService_GetAllOrder(t *testing.T) {
	mockRepo, service := setupServiceTest(t)

	t.Run("Success", func(t *testing.T) {
		query := dto.OrderQuery{Page: 1, Limit: 10}
		userInfo := models.UserInfo{Username: "testuser"}
		userInfoJSON, _ := json.Marshal(userInfo)

		mockOrders := []models.Order{
			{
				ID:            1,
				TotalAmount:   1500,
				CurrentStatus: models.ORDER_STATUS_CREATED,
				UserInfo:      userInfoJSON,
				CreatedAt:     time.Now(),
			},
		}

		mockRepo.On("GetAllOrder", query).Return(mockOrders, int64(1), nil).Once()

		orders, total, err := service.GetAllOrder(query)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, orders, 1)
		mockRepo.AssertExpectations(t)
	})
}
