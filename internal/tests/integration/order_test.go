package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"main/internal/dto"
	"main/internal/models"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationCreateOrder(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		apiKey         string
		expectedStatus int
		expectError    bool
	}{
		{
			name: "create order amount zero",
			body: dto.OrderRequest{
				TotalAmount:     0,
				Username:        "integration_user",
				UserPhone:       "0901234567",
				ShippingAddress: "123 Test Street, HCM City",
			},
			apiKey:         customerAPIKey,
			expectedStatus: 201, // Note: The handler currently doesn't validate TotalAmount > 0, so it returns 201. Added to increase coverage in controller parsing
			expectError:    false,
		},
		{
			name:           "invalid json",
			body:           "{invalid-json",
			apiKey:         customerAPIKey,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "unauthenticated",
			body: dto.OrderRequest{
				TotalAmount: 1000,
			},
			expectedStatus: 401,
			expectError:    true,
		},
		{
			name: "wrong role - admin",
			body: dto.OrderRequest{
				TotalAmount: 1000,
			},
			apiKey:         adminAPIKey,
			expectedStatus: 403,
			expectError:    true,
		},
		{
			name: "wrong role - driver",
			body: dto.OrderRequest{
				TotalAmount: 1000,
			},
			apiKey:         driverAPIKey,
			expectedStatus: 403,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanOrders()

			var bodyBytes []byte

			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(
				"POST",
				"/api/v1/orders",
				bytes.NewBuffer(bodyBytes),
			)

			req.Header.Set("Content-Type", "application/json")

			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			resp, err := app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if !tt.expectError {
				var order models.Order

				err = db.First(&order).Error
				require.NoError(t, err)

				if reqBody, ok := tt.body.(dto.OrderRequest); ok {
					assert.Equal(t, reqBody.TotalAmount, order.TotalAmount)
					assert.Equal(t, models.ORDER_STATUS_CREATED, order.CurrentStatus)

					var userInfo models.UserInfo
					json.Unmarshal(order.UserInfo, &userInfo)

					assert.Equal(t, reqBody.Username, userInfo.Username)
					assert.Equal(t, reqBody.UserPhone, userInfo.UserPhone)
					assert.Equal(t, reqBody.ShippingAddress, userInfo.ShippingAddress)
				}
			}
		})
	}
}

func TestIntegrationGetAllOrders(t *testing.T) {
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	tests := []struct {
		name           string
		query          string
		apiKey         string
		seedOrders     []models.Order
		expectedStatus int
		expectError    bool
		expectedTotal  int64
		expectedLength int
		expectedUser   string
	}{
		{
			name:   "get all orders success",
			query:  "/api/v1/orders?page=1&limit=10",
			apiKey: adminAPIKey,
			seedOrders: []models.Order{
				{
					TotalAmount:   1000,
					CurrentStatus: models.ORDER_STATUS_CREATED,
					UserInfo:      mustMarshalUserInfo("alice", "", ""),
					CreatedAt:     today.Add(-3 * time.Minute),
					UpdatedAt:     today,
				},
				{
					TotalAmount:   2000,
					CurrentStatus: models.ORDER_STATUS_PAID,
					UserInfo:      mustMarshalUserInfo("bob", "", ""),
					CreatedAt:     today.Add(-2 * time.Minute),
					UpdatedAt:     today,
				},
				{
					TotalAmount:   3000,
					CurrentStatus: models.ORDER_STATUS_DELIVERED,
					UserInfo:      mustMarshalUserInfo("carol", "", ""),
					CreatedAt:     today.Add(-1 * time.Minute),
					UpdatedAt:     today,
				},
			},
			expectedStatus: 200,
			expectError:    false,
			expectedTotal:  3,
			expectedLength: 3,
			expectedUser:   "carol",
		},
		{
			name:   "filter by status",
			query:  "/api/v1/orders?status=paid&page=1&limit=10",
			apiKey: adminAPIKey,
			seedOrders: []models.Order{
				{
					TotalAmount:   1000,
					CurrentStatus: models.ORDER_STATUS_CREATED,
					UserInfo:      mustMarshalUserInfo("u1", "", ""),
					CreatedAt:     today,
					UpdatedAt:     today,
				},
				{
					TotalAmount:   2000,
					CurrentStatus: models.ORDER_STATUS_PAID,
					UserInfo:      mustMarshalUserInfo("u2", "", ""),
					CreatedAt:     today,
					UpdatedAt:     today,
				},
				{
					TotalAmount:   3000,
					CurrentStatus: models.ORDER_STATUS_PAID,
					UserInfo:      mustMarshalUserInfo("u3", "", ""),
					CreatedAt:     today,
					UpdatedAt:     today,
				},
			},
			expectedStatus: 200,
			expectError:    false,
			expectedTotal:  2,
			expectedLength: 2,
		},
		{
			name:   "filter by date",
			query:  fmt.Sprintf("/api/v1/orders?date=%s&page=1&limit=10", today.Format("2006-01-02")),
			apiKey: adminAPIKey,
			seedOrders: []models.Order{
				{
					TotalAmount:   1000,
					CurrentStatus: models.ORDER_STATUS_CREATED,
					UserInfo:      mustMarshalUserInfo("today_user", "", ""),
					CreatedAt:     today,
					UpdatedAt:     today,
				},
				{
					TotalAmount:   2000,
					CurrentStatus: models.ORDER_STATUS_CREATED,
					UserInfo:      mustMarshalUserInfo("yesterday_user", "", ""),
					CreatedAt:     yesterday,
					UpdatedAt:     yesterday,
				},
			},
			expectedStatus: 200,
			expectError:    false,
			expectedTotal:  1,
			expectedLength: 1,
			expectedUser:   "today_user",
		},
		{
			name:   "customer access allowed",
			query:  "/api/v1/orders?page=1&limit=10",
			apiKey: customerAPIKey,
			seedOrders: []models.Order{
				{TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, UserInfo: mustMarshalUserInfo("u1", "", ""), CreatedAt: today, UpdatedAt: today},
			},
			expectedStatus: 200,
			expectError:    false,
			expectedTotal:  1,
			expectedLength: 1,
		},
		{
			name:   "pagination",
			query:  "/api/v1/orders?page=1&limit=2",
			apiKey: adminAPIKey,
			seedOrders: []models.Order{
				{TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, UserInfo: mustMarshalUserInfo("u1", "", ""), CreatedAt: today, UpdatedAt: today},
				{TotalAmount: 2000, CurrentStatus: models.ORDER_STATUS_CREATED, UserInfo: mustMarshalUserInfo("u2", "", ""), CreatedAt: today, UpdatedAt: today},
				{TotalAmount: 3000, CurrentStatus: models.ORDER_STATUS_CREATED, UserInfo: mustMarshalUserInfo("u3", "", ""), CreatedAt: today, UpdatedAt: today},
				{TotalAmount: 4000, CurrentStatus: models.ORDER_STATUS_CREATED, UserInfo: mustMarshalUserInfo("u4", "", ""), CreatedAt: today, UpdatedAt: today},
				{TotalAmount: 5000, CurrentStatus: models.ORDER_STATUS_CREATED, UserInfo: mustMarshalUserInfo("u5", "", ""), CreatedAt: today, UpdatedAt: today},
			},
			expectedStatus: 200,
			expectError:    false,
			expectedTotal:  5,
			expectedLength: 2,
		},
		{
			name:           "unauthenticated",
			query:          "/api/v1/orders",
			expectedStatus: 401,
			expectError:    true,
		},
		{
			name:   "driver access allowed",
			query:  "/api/v1/orders?page=1&limit=2",
			apiKey: driverAPIKey,
			seedOrders: []models.Order{
				{TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, UserInfo: mustMarshalUserInfo("u1", "", ""), CreatedAt: today, UpdatedAt: today},
			},
			expectedStatus: 200,
			expectError:    false,
			expectedTotal:  1,
			expectedLength: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanOrders()

			if len(tt.seedOrders) > 0 {
				db.Create(&tt.seedOrders)
			}

			req := httptest.NewRequest("GET", tt.query, nil)

			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			resp, err := app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if !tt.expectError {
				var body struct {
					Status     int                `json:"status"`
					Data       []dto.OrderReponse `json:"data"`
					Pagination struct {
						TotalItems int64 `json:"total_items"`
					} `json:"pagination"`
				}

				err = json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedTotal, body.Pagination.TotalItems)
				assert.Len(t, body.Data, tt.expectedLength)

				if tt.expectedUser != "" {
					assert.Equal(t, tt.expectedUser, body.Data[0].Username)
				}
			}
		})
	}
}

func TestIntegrationGetOrderDetail(t *testing.T) {
	tests := []struct {
		name           string
		order          models.Order
		orderID        string
		apiKey         string
		expectedStatus int
		expectError    bool
	}{
		{
			name: "get order detail success",
			order: models.Order{
				TotalAmount:   5000,
				CurrentStatus: models.ORDER_STATUS_CREATED,
				UserInfo: mustMarshalUserInfo(
					"kien",
					"0901234567",
					"HCM City",
				),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			orderID:        "1",
			apiKey:         adminAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name:           "invalid order id",
			orderID:        "abc",
			apiKey:         adminAPIKey,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:           "order not found",
			orderID:        "999",
			apiKey:         adminAPIKey,
			expectedStatus: 404,
			expectError:    true,
		},
		{
			name:           "unauthenticated",
			orderID:        "1",
			expectedStatus: 401,
			expectError:    true,
		},
		{
			name: "customer access allowed",
			order: models.Order{
				TotalAmount:   5000,
				CurrentStatus: models.ORDER_STATUS_CREATED,
				UserInfo: mustMarshalUserInfo(
					"kien",
					"0901234567",
					"HCM City",
				),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			orderID:        "1",
			apiKey:         customerAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "driver access allowed",
			order: models.Order{
				TotalAmount:   5000,
				CurrentStatus: models.ORDER_STATUS_CREATED,
				UserInfo: mustMarshalUserInfo(
					"kien",
					"0901234567",
					"HCM City",
				),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			orderID:        "1",
			apiKey:         driverAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanOrders()

			if tt.order.TotalAmount != 0 {
				db.Create(&tt.order)
			}

			req := httptest.NewRequest(
				"GET",
				"/api/v1/orders/"+tt.orderID,
				nil,
			)

			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			resp, err := app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if !tt.expectError {
				var body struct {
					Status int              `json:"status"`
					Data   dto.OrderReponse `json:"data"`
				}

				err = json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				assert.Equal(t, 200, body.Status)
				assert.Equal(t, int64(5000), body.Data.TotalAmount)
				assert.Equal(t, "kien", body.Data.Username)
				assert.Equal(t, models.ORDER_STATUS_CREATED, body.Data.Status)
			}
		})
	}
}

func TestIntegrationUpdateOrderStatus(t *testing.T) {
	tests := []struct {
		name           string
		order          models.Order
		orderID        string
		body           interface{}
		apiKey         string
		expectedStatus int
		expectError    bool
	}{
		{
			name: "update status success",
			order: models.Order{
				TotalAmount:   1000,
				CurrentStatus: models.ORDER_STATUS_CREATED,
				UserInfo:      mustMarshalUserInfo("kien", "", ""),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			orderID: "1",
			body: dto.UpdateStatusRequest{
				Status: models.ORDER_STATUS_PAID,
			},
			apiKey:         adminAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "update status packed",
			order: models.Order{
				TotalAmount:   1000,
				CurrentStatus: models.ORDER_STATUS_PAID,
				UserInfo:      mustMarshalUserInfo("kien", "", ""),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			orderID: "1",
			body: dto.UpdateStatusRequest{
				Status: models.ORDER_STATUS_PACKED,
			},
			apiKey:         adminAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "update status shipped",
			order: models.Order{
				TotalAmount:   1000,
				CurrentStatus: models.ORDER_STATUS_PACKED,
				UserInfo:      mustMarshalUserInfo("kien", "", ""),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			orderID: "1",
			body: dto.UpdateStatusRequest{
				Status: models.ORDER_STATUS_SHIPPED,
			},
			apiKey:         adminAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "update status delivered",
			order: models.Order{
				TotalAmount:   1000,
				CurrentStatus: models.ORDER_STATUS_SHIPPED,
				UserInfo:      mustMarshalUserInfo("kien", "", ""),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			orderID: "1",
			body: dto.UpdateStatusRequest{
				Status: models.ORDER_STATUS_DELIVERED,
			},
			apiKey:         adminAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "update status cancelled from created",
			order: models.Order{
				TotalAmount:   1000,
				CurrentStatus: models.ORDER_STATUS_CREATED,
				UserInfo:      mustMarshalUserInfo("kien", "", ""),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			orderID: "1",
			body: dto.UpdateStatusRequest{
				Status: models.ORDER_STATUS_CANCELLED,
			},
			apiKey:         customerAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "update status refunded from delivered",
			order: models.Order{
				TotalAmount:   1000,
				CurrentStatus: models.ORDER_STATUS_DELIVERED,
				UserInfo:      mustMarshalUserInfo("kien", "", ""),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			orderID: "1",
			body: dto.UpdateStatusRequest{
				Status: models.ORDER_STATUS_REFUNDED,
			},
			apiKey:         adminAPIKey,
			expectedStatus: 400, // Invalid transition from delivered to refunded (based on models.IsValidTransition)
			expectError:    true,
		},
		{
			name: "wrong role - driver",
			order: models.Order{
				TotalAmount:   1000,
				CurrentStatus: models.ORDER_STATUS_CREATED,
				UserInfo:      mustMarshalUserInfo("kien", "", ""),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			orderID: "1",
			body: dto.UpdateStatusRequest{
				Status: models.ORDER_STATUS_PAID,
			},
			apiKey:         driverAPIKey,
			expectedStatus: 403,
			expectError:    true,
		},
		{
			name: "invalid status transition",
			order: models.Order{
				TotalAmount:   1000,
				CurrentStatus: models.ORDER_STATUS_DELIVERED,
				UserInfo:      mustMarshalUserInfo("kien", "", ""),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			orderID: "1",
			body: dto.UpdateStatusRequest{
				Status: models.ORDER_STATUS_CREATED,
			},
			apiKey:         adminAPIKey,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:    "invalid order id",
			orderID: "abc",
			body: dto.UpdateStatusRequest{
				Status: models.ORDER_STATUS_PAID,
			},
			apiKey:         adminAPIKey,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:    "empty status",
			orderID: "1",
			body:    dto.UpdateStatusRequest{},
			apiKey:  adminAPIKey,

			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:    "order not found",
			orderID: "999",
			body: dto.UpdateStatusRequest{
				Status: models.ORDER_STATUS_PAID,
			},
			apiKey:         adminAPIKey,
			expectedStatus: 404,
			expectError:    true,
		},
		{
			name:    "invalid json",
			orderID: "1",
			body:    "{invalid-json",
			apiKey:  adminAPIKey,

			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanOrders()

			if tt.order.TotalAmount != 0 {
				db.Create(&tt.order)
			}

			var bodyBytes []byte

			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(
				"PATCH",
				"/api/v1/orders/"+tt.orderID+"/status",
				bytes.NewBuffer(bodyBytes),
			)

			req.Header.Set("Content-Type", "application/json")

			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			resp, err := app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if !tt.expectError {
				var updatedOrder models.Order

				err = db.First(&updatedOrder, 1).Error
				require.NoError(t, err)

				var expectedStatus models.OrderStatus
				if reqBody, ok := tt.body.(dto.UpdateStatusRequest); ok {
					expectedStatus = reqBody.Status
				} else {
					expectedStatus = models.ORDER_STATUS_PAID // Fallback
				}

				assert.Equal(t, expectedStatus, updatedOrder.CurrentStatus)
			}
		})
	}
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
