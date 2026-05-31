package services_test

import (
	"encoding/json"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/services"
	"main/internal/tests/unit/mocks"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrderServiceCreateOrder(t *testing.T) {
	testCases := []struct {
		name           string
		input          dto.OrderRequest
		setupMock      func(*mocks.OrderRepository)
		expectError    bool
		expectedResult *dto.CreateOrderResponse
	}{
		{
			name: "Success",
			input: dto.OrderRequest{
				TotalAmount:     2000,
				Username:        "testuser",
				UserPhone:       "0987654321",
				ShippingAddress: "123 Test Street, HCM City",
			},
			setupMock: func(mockRepo *mocks.OrderRepository) {
				req := dto.OrderRequest{
					TotalAmount:     2000,
					Username:        "testuser",
					UserPhone:       "0987654321",
					ShippingAddress: "123 Test Street, HCM City",
				}

				mockRepo.
					On("CreateOrder", buildExpectedOrder(req)).
					Return(&models.Order{
						ID:            1,
						TotalAmount:   2000,
						CurrentStatus: models.ORDER_STATUS_CREATED,
						UserInfo:      marshalUserInfo(req),
					}, nil).
					Once()
			},
			expectError: false,
			expectedResult: &dto.CreateOrderResponse{
				ID:          1,
				TotalAmount: 2000,
				Status:      models.ORDER_STATUS_CREATED,
			},
		},
		{
			name: "Create error",
			input: dto.OrderRequest{
				TotalAmount: 1000,
				Username:    "failuser",
			},
			setupMock: func(mockRepo *mocks.OrderRepository) {
				req := dto.OrderRequest{
					TotalAmount: 1000,
					Username:    "failuser",
				}

				mockRepo.
					On("CreateOrder", buildExpectedOrder(req)).
					Return(nil, assert.AnError).
					Once()
			},
			expectError:    true,
			expectedResult: nil,
		},
		{
			name: "Invalid max total amount",
			input: dto.OrderRequest{
				TotalAmount: math.MaxInt64,
				Username:    "bigspender",
			},
			setupMock: func(mockRepo *mocks.OrderRepository) {
				req := dto.OrderRequest{
					TotalAmount: math.MaxInt64,
					Username:    "bigspender",
				}

				mockRepo.
					On("CreateOrder", buildExpectedOrder(req)).
					Return(&models.Order{
						ID:            4,
						TotalAmount:   math.MaxInt64,
						CurrentStatus: models.ORDER_STATUS_CREATED,
						UserInfo:      marshalUserInfo(req),
					}, nil).
					Once()
			},
			expectError: false,
			expectedResult: &dto.CreateOrderResponse{
				ID:          4,
				TotalAmount: math.MaxInt64,
				Status:      models.ORDER_STATUS_CREATED,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewOrderRepository(t)
			service := services.NewOrderService(mockRepo)

			tc.setupMock(mockRepo)

			result, err := service.CreateOrder(tc.input)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				assert.Equal(t, tc.expectedResult.ID, result.ID)
				assert.Equal(t, tc.expectedResult.TotalAmount, result.TotalAmount)
				assert.Equal(t, tc.expectedResult.Status, result.Status)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestOrderServiceGetAllOrder(t *testing.T) {
	testCases := []struct {
		name           string
		input          dto.OrderQuery
		setupMock      func(*mocks.OrderRepository)
		expectError    bool
		expectedOrders []dto.OrderReponse
		expectedTotal  int64
	}{
		{
			name:  "Success",
			input: dto.OrderQuery{PageNumber: 1, LimitItems: 10},
			setupMock: func(mockRepo *mocks.OrderRepository) {
				mockOrders := []models.Order{
					{
						ID:            1,
						TotalAmount:   1500,
						CurrentStatus: models.ORDER_STATUS_CREATED,
						UserInfo:      mustMarshalUserInfo("alice", "0900000001", "Addr A"),
						CreatedAt:     time.Now(),
					},
					{
						ID:            2,
						TotalAmount:   2500,
						CurrentStatus: models.ORDER_STATUS_PAID,
						UserInfo:      mustMarshalUserInfo("bob", "0900000002", "Addr B"),
						CreatedAt:     time.Now(),
					},
				}

				mockRepo.
					On("GetAllOrder", dto.OrderQuery{PageNumber: 1, LimitItems: 10}).
					Return(mockOrders, int64(2), nil).
					Once()
			},
			expectError:   false,
			expectedTotal: 2,
			expectedOrders: []dto.OrderReponse{
				{
					ID:              1,
					Username:        "alice",
					UserPhone:       "0900000001",
					ShippingAddress: "Addr A",
					TotalAmount:     1500,
					Status:          models.ORDER_STATUS_CREATED,
				},
				{
					ID:              2,
					Username:        "bob",
					UserPhone:       "0900000002",
					ShippingAddress: "Addr B",
					TotalAmount:     2500,
					Status:          models.ORDER_STATUS_PAID,
				},
			},
		},
		{
			name:  "Get data error",
			input: dto.OrderQuery{PageNumber: 1, LimitItems: 10},
			setupMock: func(mockRepo *mocks.OrderRepository) {
				mockRepo.
					On("GetAllOrder", dto.OrderQuery{PageNumber: 1, LimitItems: 10}).
					Return(nil, int64(0), assert.AnError).
					Once()
			},
			expectError:    true,
			expectedOrders: nil,
			expectedTotal:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewOrderRepository(t)
			service := services.NewOrderService(mockRepo)

			tc.setupMock(mockRepo)

			orders, total, err := service.GetAllOrder(tc.input)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				assert.Equal(t, tc.expectedTotal, total)
				assert.Len(t, orders, len(tc.expectedOrders))

				for i := range tc.expectedOrders {
					assert.Equal(t, tc.expectedOrders[i].Username, orders[i].Username)
					assert.Equal(t, tc.expectedOrders[i].UserPhone, orders[i].UserPhone)
					assert.Equal(t, tc.expectedOrders[i].ShippingAddress, orders[i].ShippingAddress)
					assert.Equal(t, tc.expectedOrders[i].TotalAmount, orders[i].TotalAmount)
					assert.Equal(t, tc.expectedOrders[i].Status, orders[i].Status)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestOrderServiceGetOrder(t *testing.T) {
	userInfo := models.UserInfo{
		Username:        "testuser",
		UserPhone:       "0123456789",
		ShippingAddress: "123 Test Street",
	}

	userInfoJSON, _ := json.Marshal(userInfo)

	testCases := []struct {
		name           string
		orderID        int64
		setupMock      func(*mocks.OrderRepository)
		expectError    bool
		expectedResult *dto.OrderReponse
	}{
		{
			name:    "Success",
			orderID: 123,
			setupMock: func(mockRepo *mocks.OrderRepository) {
				mockRepo.
					On("GetOrderDetail", int64(123)).
					Return(&models.Order{
						ID:            123,
						TotalAmount:   2500000,
						CurrentStatus: models.ORDER_STATUS_CREATED,
						UserInfo:      userInfoJSON,
						CreatedAt:     time.Now(),
					}, nil).
					Once()
			},
			expectError: false,
			expectedResult: &dto.OrderReponse{
				ID:              123,
				Username:        "testuser",
				UserPhone:       "0123456789",
				ShippingAddress: "123 Test Street",
				TotalAmount:     2500000,
				Status:          models.ORDER_STATUS_CREATED,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewOrderRepository(t)
			service := services.NewOrderService(mockRepo)

			tc.setupMock(mockRepo)

			result, err := service.GetOrder(tc.orderID)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				assert.Equal(t, tc.expectedResult.ID, result.ID)
				assert.Equal(t, tc.expectedResult.Username, result.Username)
				assert.Equal(t, tc.expectedResult.UserPhone, result.UserPhone)
				assert.Equal(t, tc.expectedResult.ShippingAddress, result.ShippingAddress)
				assert.Equal(t, tc.expectedResult.TotalAmount, result.TotalAmount)
				assert.Equal(t, tc.expectedResult.Status, result.Status)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestOrderServiceUpdateOrderStatus(t *testing.T) {
	testCases := []struct {
		name           string
		orderID        int64
		status         string
		setupMock      func(*mocks.OrderRepository)
		expectError    bool
		expectedResult *models.Order
	}{
		{
			name:    "Success",
			orderID: 123,
			status:  "paid",
			setupMock: func(mockRepo *mocks.OrderRepository) {
				mockRepo.
					On("GetOrderDetail", int64(123)).
					Return(&models.Order{
						ID:            123,
						CurrentStatus: models.ORDER_STATUS_CREATED,
					}, nil).
					Once()

				mockRepo.
					On("UpdateOrderStatus", int64(123), "paid").
					Return(&models.Order{
						ID:            123,
						TotalAmount:   2500000,
						CurrentStatus: models.ORDER_STATUS_PAID,
						CreatedAt:     time.Now(),
					}, nil).
					Once()
			},
			expectError: false,
			expectedResult: &models.Order{
				ID:            123,
				CurrentStatus: models.ORDER_STATUS_PAID,
			},
		},
		{
			name:    "Invalid transition",
			orderID: 888,
			status:  "shipped",
			setupMock: func(mockRepo *mocks.OrderRepository) {
				mockRepo.
					On("GetOrderDetail", int64(888)).
					Return(&models.Order{
						ID:            888,
						CurrentStatus: models.ORDER_STATUS_CREATED,
					}, nil).
					Once()
			},
			expectError:    true,
			expectedResult: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewOrderRepository(t)
			service := services.NewOrderService(mockRepo)

			tc.setupMock(mockRepo)

			result, err := service.UpdateOrderStatus(tc.orderID, tc.status)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				assert.Equal(t, tc.expectedResult.ID, result.ID)
				assert.Equal(t, tc.expectedResult.CurrentStatus, result.CurrentStatus)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func buildExpectedOrder(req dto.OrderRequest) models.Order {
	userInfo := models.UserInfo{
		Username:        req.Username,
		UserPhone:       req.UserPhone,
		ShippingAddress: req.ShippingAddress,
	}
	userInfoJSON, _ := json.Marshal(userInfo)

	return models.Order{
		UserInfo:      userInfoJSON,
		TotalAmount:   req.TotalAmount,
		CurrentStatus: models.ORDER_STATUS_CREATED,
	}
}

func marshalUserInfo(req dto.OrderRequest) []byte {
	userInfo := models.UserInfo{
		Username:        req.Username,
		UserPhone:       req.UserPhone,
		ShippingAddress: req.ShippingAddress,
	}
	b, _ := json.Marshal(userInfo)
	return b
}

func mustMarshalUserInfo(username, phone, address string) []byte {
	info := models.UserInfo{
		Username:        username,
		UserPhone:       phone,
		ShippingAddress: address,
	}
	b, _ := json.Marshal(info)
	return b
}
