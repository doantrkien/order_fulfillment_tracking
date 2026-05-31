package handlers_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"main/errs"
	"main/internal/dto"
	"main/internal/handlers"
	"main/internal/models"
	"main/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"main/internal/tests/unit/mocks"
)

func TestOrderHandlerCreateOrder(t *testing.T) {
	tests := []struct {
		name           string
		input          dto.OrderRequest
		setupMock      func(*mocks.OrderRepository)
		expectError    bool
		expectedResult *models.Order
	}{
		{
			name: "Success",
			input: dto.OrderRequest{
				TotalAmount:     1000,
				Username:        "testuser",
				UserPhone:       "0123456789",
				ShippingAddress: "Test Address",
			},
			setupMock: func(m *mocks.OrderRepository) {
				m.On("CreateOrder", mock.Anything).Return(&models.Order{ID: 1}, nil).Once()
			},
			expectError:    false,
			expectedResult: &models.Order{ID: 1},
		},
		{
			name: "Invalid JSON",
			input: dto.OrderRequest{
				TotalAmount: 1000,
			},
			setupMock:   nil,
			expectError: true,
		},
		{
			name:        "Empty body",
			input:       dto.OrderRequest{},
			setupMock:   nil,
			expectError: true,
		},
		{
			name: "Service error",
			input: dto.OrderRequest{
				TotalAmount:     2000,
				Username:        "erroruser",
				UserPhone:       "0123456789",
				ShippingAddress: "Test Address",
			},
			setupMock: func(m *mocks.OrderRepository) {
				m.On("CreateOrder", mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectError: true,
		},
		{
			name: "Missing Content-Type header",
			input: dto.OrderRequest{
				TotalAmount: 3000,
			},
			setupMock:   nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			mockRepo := mocks.NewOrderRepository(t)
			svc := services.NewOrderService(mockRepo)
			handler := handlers.NewOrderHandler(svc)
			app.Post("/orders", handler.CreateOrder)

			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}

			var reqBody []byte
			if tc.name == "Invalid JSON" {
				reqBody = []byte("{invalid-json}")
			} else if tc.name == "Empty body" {
				reqBody = []byte("")
			} else {
				reqBody, _ = json.Marshal(tc.input)
			}

			req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
			if tc.name != "Missing Content-Type header" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := app.Test(req)
			assert.NoError(t, err)

			if tc.expectError {
				if tc.name == "Invalid JSON" || tc.name == "Missing Content-Type header" || tc.name == "Empty body" {
					assert.Equal(t, 400, resp.StatusCode)
				} else {
					assert.Equal(t, 500, resp.StatusCode)
				}
			} else {
				assert.Equal(t, 201, resp.StatusCode)
			}
		})
	}
}

func TestOrderHandlerGetAllOrder(t *testing.T) {
	mockTime := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		query     string
		setupMock func(*mocks.OrderRepository)
		wantCode  int
		validate  func(*testing.T, *http.Response)
	}{
		{
			name:  "Success default pagination",
			query: "",
			setupMock: func(m *mocks.OrderRepository) {
				mockOrders := []models.Order{
					{
						ID:            1,
						TotalAmount:   1000,
						CurrentStatus: models.ORDER_STATUS_CREATED,
						UserInfo:      []byte(`{"username":"testuser"}`),
						CreatedAt:     mockTime,
					},
				}
				m.On("GetAllOrder", dto.OrderQuery{PageNumber: 1, LimitItems: 10}, "admin", int64(1)).
					Return(mockOrders, int64(1), nil).Once()
			},
			wantCode: 200,
			validate: func(t *testing.T, resp *http.Response) {
				var body struct {
					Data       []dto.OrderReponse `json:"data"`
					Pagination struct {
						Page  int   `json:"current_page"`
						Limit int   `json:"limit_items"`
						Total int64 `json:"total_items"`
					} `json:"pagination"`
				}
				json.NewDecoder(resp.Body).Decode(&body)
				assert.Equal(t, 1, body.Pagination.Page)
				assert.Equal(t, 10, body.Pagination.Limit)
				assert.Equal(t, int64(1), body.Pagination.Total)
				assert.Len(t, body.Data, 1)
			},
		},
		{
			name:  "Success",
			query: "?page=2&limit=5",
			setupMock: func(m *mocks.OrderRepository) {
				mockOrders := []models.Order{
					{
						ID:            6,
						TotalAmount:   600,
						CurrentStatus: models.ORDER_STATUS_PAID,
						UserInfo:      []byte(`{"username":"testuser"}`),
						CreatedAt:     mockTime,
					},
				}
				m.On("GetAllOrder", dto.OrderQuery{PageNumber: 2, LimitItems: 5}, "admin", int64(1)).
					Return(mockOrders, int64(8), nil).Once()
			},
			wantCode: 200,
			validate: func(t *testing.T, resp *http.Response) {
				var body struct {
					Pagination struct {
						Page       int   `json:"current_page"`
						Limit      int   `json:"limit_items"`
						TotalItems int64 `json:"total_items"`
					} `json:"pagination"`
				}
				json.NewDecoder(resp.Body).Decode(&body)
				assert.Equal(t, 2, body.Pagination.Page)
				// Calculate total pages: ceil(total_items / limit_items)
				expectedTotalPages := (int(body.Pagination.TotalItems) + body.Pagination.Limit - 1) / body.Pagination.Limit
				assert.Equal(t, 2, expectedTotalPages)
			},
		},
		{
			name:  "Service error",
			query: "",
			setupMock: func(m *mocks.OrderRepository) {
				m.On("GetAllOrder", dto.OrderQuery{PageNumber: 1, LimitItems: 10}, "admin", int64(1)).Return(nil, int64(0), assert.AnError).Once()
			},
			wantCode: 500,
		},
		{
			name:  "With status filter",
			query: "?status=paid&page=1&limit=10",
			setupMock: func(m *mocks.OrderRepository) {
				mockOrders := []models.Order{
					{
						ID:            2,
						CurrentStatus: models.ORDER_STATUS_PAID,
					},
				}
				m.On("GetAllOrder", dto.OrderQuery{PageNumber: 1, LimitItems: 10, Status: "paid"}, "admin", int64(1)).
					Return(mockOrders, int64(1), nil).Once()
			},
			wantCode: 200,
		},
		{
			name:  "With date filter",
			query: "?date=2026-05-19&page=1&limit=10",
			setupMock: func(m *mocks.OrderRepository) {
				mockOrders := []models.Order{
					{
						ID: 3,
					},
				}
				m.On("GetAllOrder", dto.OrderQuery{PageNumber: 1, LimitItems: 10, Date: "2026-05-19"}, "admin", int64(1)).
					Return(mockOrders, int64(1), nil).Once()
			},
			wantCode: 200,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c fiber.Ctx) error {
				c.Locals("role", "admin")
				c.Locals("user_id", int64(1))
				return c.Next()
			})
			mockRepo := mocks.NewOrderRepository(t)
			svc := services.NewOrderService(mockRepo)
			handler := handlers.NewOrderHandler(svc)
			app.Get("/orders", handler.GetAllOrder)

			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}

			req := httptest.NewRequest("GET", "/orders"+tc.query, nil)
			resp, _ := app.Test(req)
			assert.Equal(t, tc.wantCode, resp.StatusCode)

			if tc.validate != nil {
				tc.validate(t, resp)
			}
		})
	}
}

func TestOrderHandlerGetOrderDetail(t *testing.T) {
	tests := []struct {
		name           string
		order          models.Order
		orderID        string
		token          string
		expectedStatus int
		expectError    bool
		setupMock      func(*mocks.OrderRepository)
	}{
		{
			name: "Success",
			order: models.Order{
				ID:            1,
				TotalAmount:   1000,
				CurrentStatus: models.ORDER_STATUS_CREATED,
				UserInfo:      []byte(`{"username":"kien","user_phone":"0123456789","shipping_address":"HCM"}`),
				CreatedAt:     time.Now(),
			},
			orderID:        "1",
			expectedStatus: 200,
			expectError:    false,
			setupMock: func(m *mocks.OrderRepository) {
				m.On("GetOrderDetail", int64(1), "admin", int64(1)).
					Return(&models.Order{
						ID:            1,
						TotalAmount:   1000,
						CurrentStatus: models.ORDER_STATUS_CREATED,
						UserInfo:      []byte(`{"username":"kien","user_phone":"0123456789","shipping_address":"HCM"}`),
						CreatedAt:     time.Now(),
					}, nil).
					Once()
			},
		},
		{
			name:           "Invalid order id",
			orderID:        "abc",
			expectedStatus: 400,
			expectError:    true,
		},
		// {
		// 	name:           "Order not found",
		// 	orderID:        "999",
		// 	expectedStatus: 404,
		// 	expectError:    true,
		// 	setupMock: func(m *mocks.OrderRepository) {
		// 		m.On("GetOrderDetail", int64(999)).
		// 			Return(nil, nil).
		// 			Once()
		// 	},
		// },
		{
			name:           "Service error",
			orderID:        "2",
			expectedStatus: 500,
			expectError:    true,
			setupMock: func(m *mocks.OrderRepository) {
				m.On("GetOrderDetail", int64(2), "admin", int64(1)).
					Return(nil, assert.AnError).
					Once()
			},
		},
		{
			name:           "Unauthorized access",
			orderID:        "1",
			expectedStatus: 403,
			expectError:    true,
			setupMock: func(m *mocks.OrderRepository) {
				m.On("GetOrderDetail", int64(1), "admin", int64(1)).
					Return(nil, errs.ERR_UNAUTHORIZED).
					Once()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c fiber.Ctx) error {
				c.Locals("role", "admin")
				c.Locals("user_id", int64(1))
				return c.Next()
			})

			mockRepo := mocks.NewOrderRepository(t)
			svc := services.NewOrderService(mockRepo)
			handler := handlers.NewOrderHandler(svc)

			app.Get("/orders/:id", handler.GetOrderDetail)

			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}

			req := httptest.NewRequest(
				"GET",
				"/orders/"+tc.orderID,
				nil,
			)

			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestOrderHandlerUpdateOrderStatus(t *testing.T) {
	tests := []struct {
		name           string
		order          models.Order
		orderID        string
		token          string
		expectedStatus int
		expectError    bool
		body           interface{}
		setupMock      func(*mocks.OrderRepository)
	}{
		{
			name: "Success",
			order: models.Order{
				ID:            1,
				CurrentStatus: models.ORDER_STATUS_CREATED,
			},
			orderID: "1",
			body: dto.UpdateStatusRequest{
				Status: "paid",
			},
			expectedStatus: 200,
			expectError:    false,
			setupMock: func(m *mocks.OrderRepository) {
				// m.On("UpdateOrderStatus", int64(1), "paid", "email").
				m.On("UpdateOrderStatus", int64(1), "paid", "1").
					Return(&models.Order{
						ID:            1,
						CurrentStatus: models.ORDER_STATUS_PAID,
					}, nil).
					Once()
			},
		},
		{
			name:           "Invalid order id",
			orderID:        "abc",
			expectedStatus: 400,
			expectError:    true,
			body: dto.UpdateStatusRequest{
				Status: "paid",
			},
		},
		{
			name:           "Invalid json",
			orderID:        "1",
			expectedStatus: 400,
			expectError:    true,
			body:           "{invalid-json",
		},
		{
			name:           "Empty status",
			orderID:        "1",
			expectedStatus: 400,
			expectError:    true,
			body: dto.UpdateStatusRequest{
				Status: "",
			},
		},
		// {
		// 	name:           "Order not found",
		// 	orderID:        "999",
		// 	expectedStatus: 404,
		// 	expectError:    true,
		// 	body: dto.UpdateStatusRequest{
		// 		Status: "paid",
		// 	},
		// 	setupMock: func(m *mocks.OrderRepository) {
		// 		m.On("GetOrderDetail", int64(999)).
		// 			Return(nil, nil).
		// 			Once()
		// 	},
		// },
		{
			name:           "Invalid transition",
			orderID:        "1",
			expectedStatus: 400,
			expectError:    true,
			body: dto.UpdateStatusRequest{
				Status: "delivered",
			},
			setupMock: func(m *mocks.OrderRepository) {
				m.On("UpdateOrderStatus", int64(1), "delivered", "1").
					Return(nil, errs.ERR_INVALID_STATUS).
					Once()
			},
		},
		{
			name:           "Update error",
			orderID:        "1",
			expectedStatus: 500,
			expectError:    true,
			body: dto.UpdateStatusRequest{
				Status: "paid",
			},
			setupMock: func(m *mocks.OrderRepository) {
				m.On("UpdateOrderStatus", int64(1), "paid", "1").
					Return(nil, assert.AnError).
					Once()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(func(c fiber.Ctx) error {
				c.Locals("email", "admin@example.com")
				c.Locals("user_id", int64(1))
				return c.Next()
			})

			mockRepo := mocks.NewOrderRepository(t)
			svc := services.NewOrderService(mockRepo)
			handler := handlers.NewOrderHandler(svc)

			app.Patch("/orders/:id/status", handler.UpdateOrderStatus)

			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}

			var bodyBytes []byte

			switch v := tc.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(
				"PATCH",
				"/orders/"+tc.orderID+"/status",
				bytes.NewBuffer(bodyBytes),
			)

			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}
