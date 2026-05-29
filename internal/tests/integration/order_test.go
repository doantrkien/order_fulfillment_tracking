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

// seedUser tạo user trong DB để dùng chung cho integration test
func seedUser(t *testing.T, username, phone, address string) models.User {
	t.Helper()
	user := models.User{
		Username: username,
		Password: "hashed_password",
		Phone:    phone,
		Address:  address,
		Role:     "customer",
	}
	err := db.Create(&user).Error
	require.NoError(t, err)
	return user
}

func TestIntegrationCreateOrder(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		apiKey         string
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Success",
			body: dto.OrderRequest{
				UserID:      1,
				TotalAmount: 10,
			},
			apiKey:         adminAPIKey,
			expectedStatus: 201,
			expectError:    false,
		},
		{
			name:           "invalid json",
			body:           "{invalid-json",
			apiKey:         adminAPIKey,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "unauthenticated",
			body: dto.OrderRequest{
				UserID:      1,
				TotalAmount: 1000,
			},
			expectedStatus: 401,
			expectError:    true,
		},
		{
			name: "Invalid Input - negative amount",
			body: dto.OrderRequest{
				UserID:      1,
				TotalAmount: -10,
			},
			apiKey:         adminAPIKey,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "wrong role - driver",
			body: dto.OrderRequest{
				UserID:      1,
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

			req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(bodyBytes))
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
					assert.Equal(t, reqBody.UserID, order.UserID)
					assert.Equal(t, models.ORDER_STATUS_CREATED, order.CurrentStatus)
				}
			}
		})
	}
}

func TestIntegrationGetAllOrders(t *testing.T) {
	today := time.Now().UTC()
	yesterday := today.AddDate(0, 0, -1)

	tests := []struct {
		name           string
		query          string
		apiKey         string
		seedOrders     func(t *testing.T) []models.Order
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
			seedOrders: func(t *testing.T) []models.Order {
				u1 := seedUser(t, "alice", "", "")
				u2 := seedUser(t, "bob", "", "")
				u3 := seedUser(t, "carol", "", "")
				return []models.Order{
					{UserID: u1.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: today.Add(-3 * time.Minute), UpdatedAt: today},
					{UserID: u2.ID, TotalAmount: 2000, CurrentStatus: models.ORDER_STATUS_PAID, CreatedAt: today.Add(-2 * time.Minute), UpdatedAt: today},
					{UserID: u3.ID, TotalAmount: 3000, CurrentStatus: models.ORDER_STATUS_DELIVERED, CreatedAt: today.Add(-1 * time.Minute), UpdatedAt: today},
				}
			},
			expectedStatus: 200,
			expectError:    false,
			expectedTotal:  3,
			expectedLength: 3,
			expectedUser:   "carol", // DESC order → carol là mới nhất
		},
		{
			name:   "filter by status",
			query:  "/api/v1/orders?status=paid&page=1&limit=10",
			apiKey: adminAPIKey,
			seedOrders: func(t *testing.T) []models.Order {
				u1 := seedUser(t, "u1", "", "")
				u2 := seedUser(t, "u2", "", "")
				u3 := seedUser(t, "u3", "", "")
				return []models.Order{
					{UserID: u1.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: today, UpdatedAt: today},
					{UserID: u2.ID, TotalAmount: 2000, CurrentStatus: models.ORDER_STATUS_PAID, CreatedAt: today, UpdatedAt: today},
					{UserID: u3.ID, TotalAmount: 3000, CurrentStatus: models.ORDER_STATUS_PAID, CreatedAt: today, UpdatedAt: today},
				}
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
			seedOrders: func(t *testing.T) []models.Order {
				u1 := seedUser(t, "today_user", "", "")
				u2 := seedUser(t, "yesterday_user", "", "")
				return []models.Order{
					{UserID: u1.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: today, UpdatedAt: today},
					{UserID: u2.ID, TotalAmount: 2000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: yesterday, UpdatedAt: yesterday},
				}
			},
			expectedStatus: 200,
			expectError:    false,
			expectedTotal:  1,
			expectedLength: 1,
			expectedUser:   "today_user",
		},
		{
			name:   "pagination",
			query:  "/api/v1/orders?page=1&limit=2",
			apiKey: adminAPIKey,
			seedOrders: func(t *testing.T) []models.Order {
				u1 := seedUser(t, "u1", "", "")
				u2 := seedUser(t, "u2", "", "")
				u3 := seedUser(t, "u3", "", "")
				u4 := seedUser(t, "u4", "", "")
				u5 := seedUser(t, "u5", "", "")
				return []models.Order{
					{UserID: u1.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: today, UpdatedAt: today},
					{UserID: u2.ID, TotalAmount: 2000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: today, UpdatedAt: today},
					{UserID: u3.ID, TotalAmount: 3000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: today, UpdatedAt: today},
					{UserID: u4.ID, TotalAmount: 4000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: today, UpdatedAt: today},
					{UserID: u5.ID, TotalAmount: 5000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: today, UpdatedAt: today},
				}
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
			seedOrders: func(t *testing.T) []models.Order {
				u1 := seedUser(t, "u1", "", "")
				return []models.Order{
					{UserID: u1.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: today, UpdatedAt: today},
				}
			},
			expectedStatus: 200,
			expectError:    false,
			expectedTotal:  1,
			expectedLength: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanAll()

			if tt.seedOrders != nil {
				orders := tt.seedOrders(t)
				db.Create(&orders)
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
		seedOrder      func(t *testing.T) models.Order
		orderID        string
		apiKey         string
		expectedStatus int
		expectError    bool
	}{
		{
			name: "get order detail success",
			seedOrder: func(t *testing.T) models.Order {
				u := seedUser(t, "kien", "0901234567", "HCM City")
				return models.Order{
					UserID:        u.ID,
					TotalAmount:   5000,
					CurrentStatus: models.ORDER_STATUS_CREATED,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
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
			name: "admin access allowed",
			seedOrder: func(t *testing.T) models.Order {
				u := seedUser(t, "kien", "0901234567", "HCM City")
				return models.Order{
					UserID:        u.ID,
					TotalAmount:   5000,
					CurrentStatus: models.ORDER_STATUS_CREATED,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
			},
			orderID:        "1",
			apiKey:         adminAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "driver access allowed",
			seedOrder: func(t *testing.T) models.Order {
				u := seedUser(t, "kien", "0901234567", "HCM City")
				return models.Order{
					UserID:        u.ID,
					TotalAmount:   5000,
					CurrentStatus: models.ORDER_STATUS_CREATED,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
			},
			orderID:        "1",
			apiKey:         driverAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanAll()

			if tt.seedOrder != nil {
				order := tt.seedOrder(t)
				db.Create(&order)
			}

			req := httptest.NewRequest("GET", "/api/v1/orders/"+tt.orderID, nil)
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
		seedOrder      func(t *testing.T) models.Order
		orderID        string
		body           interface{}
		apiKey         string
		expectedStatus int
		expectError    bool
	}{
		{
			name: "update status success - created to paid",
			seedOrder: func(t *testing.T) models.Order {
				u := seedUser(t, "kien", "", "")
				return models.Order{UserID: u.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			},
			orderID:        "1",
			body:           dto.UpdateStatusRequest{Status: models.ORDER_STATUS_PAID},
			apiKey:         adminAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "update status paid to packed",
			seedOrder: func(t *testing.T) models.Order {
				u := seedUser(t, "kien", "", "")
				return models.Order{UserID: u.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_PAID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			},
			orderID:        "1",
			body:           dto.UpdateStatusRequest{Status: models.ORDER_STATUS_PACKED},
			apiKey:         adminAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "update status packed to shipped",
			seedOrder: func(t *testing.T) models.Order {
				u := seedUser(t, "kien", "", "")
				return models.Order{UserID: u.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_PACKED, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			},
			orderID:        "1",
			body:           dto.UpdateStatusRequest{Status: models.ORDER_STATUS_SHIPPED},
			apiKey:         adminAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "update status shipped to delivered",
			seedOrder: func(t *testing.T) models.Order {
				u := seedUser(t, "kien", "", "")
				return models.Order{UserID: u.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_SHIPPED, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			},
			orderID:        "1",
			body:           dto.UpdateStatusRequest{Status: models.ORDER_STATUS_DELIVERED},
			apiKey:         adminAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "update status cancelled from created",
			seedOrder: func(t *testing.T) models.Order {
				u := seedUser(t, "kien", "", "")
				return models.Order{UserID: u.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			},
			orderID:        "1",
			body:           dto.UpdateStatusRequest{Status: models.ORDER_STATUS_CANCELLED},
			apiKey:         adminAPIKey,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "update status refunded from delivered - invalid transition",
			seedOrder: func(t *testing.T) models.Order {
				u := seedUser(t, "kien", "", "")
				return models.Order{UserID: u.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_DELIVERED, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			},
			orderID:        "1",
			body:           dto.UpdateStatusRequest{Status: models.ORDER_STATUS_REFUNDED},
			apiKey:         adminAPIKey,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "invalid status transition",
			seedOrder: func(t *testing.T) models.Order {
				u := seedUser(t, "kien", "", "")
				return models.Order{UserID: u.ID, TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_DELIVERED, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			},
			orderID:        "1",
			body:           dto.UpdateStatusRequest{Status: models.ORDER_STATUS_CREATED},
			apiKey:         adminAPIKey,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:           "invalid order id",
			orderID:        "abc",
			body:           dto.UpdateStatusRequest{Status: models.ORDER_STATUS_PAID},
			apiKey:         adminAPIKey,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:           "empty status",
			orderID:        "1",
			body:           dto.UpdateStatusRequest{},
			apiKey:         adminAPIKey,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:           "order not found",
			orderID:        "999",
			body:           dto.UpdateStatusRequest{Status: models.ORDER_STATUS_PAID},
			apiKey:         adminAPIKey,
			expectedStatus: 404,
			expectError:    true,
		},
		{
			name:           "invalid json",
			orderID:        "1",
			body:           "{invalid-json",
			apiKey:         adminAPIKey,
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanAll()

			if tt.seedOrder != nil {
				order := tt.seedOrder(t)
				db.Create(&order)
			}

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest("PATCH", "/api/v1/orders/"+tt.orderID+"/status", bytes.NewBuffer(bodyBytes))
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

				if reqBody, ok := tt.body.(dto.UpdateStatusRequest); ok {
					assert.Equal(t, reqBody.Status, updatedOrder.CurrentStatus)
				}
			}
		})
	}
}
