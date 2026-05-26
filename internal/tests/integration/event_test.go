package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"main/constant"
	"main/internal/dto"
	"main/internal/models"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedOrder(t *testing.T, totalAmount int64, status models.OrderStatus) models.Order {
	t.Helper()
	order := models.Order{
		TotalAmount:   totalAmount,
		CurrentStatus: status,
		UserInfo:      mustMarshalUserInfo("test_user", "0901234567", "123 Test St"),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err := db.Create(&order).Error
	require.NoError(t, err)
	return order
}

func TestIntegrationImportOrderEvents(t *testing.T) {
	testCases := []struct {
		name           string
		seedDB         func(t *testing.T) []dto.ImportOrderEventRequest
		rawBody        []byte // if set, overrides seedDB for invalid JSON tests
		expectedStatus int
		validate       func(t *testing.T, respBody []byte)
	}{
		{
			name: "success - valid transition",
			seedDB: func(t *testing.T) []dto.ImportOrderEventRequest {
				order := seedOrder(t, 5000, models.ORDER_STATUS_CREATED)
				return []dto.ImportOrderEventRequest{
					{OrderID: order.ID, Status: "paid", EventAt: time.Now(), UpdatedBy: "admin"},
				}
			},
			expectedStatus: 200,
			validate: func(t *testing.T, respBody []byte) {
				var body struct {
					Status  int                           `json:"status"`
					Message string                        `json:"message"`
					Data    dto.ImportOrderEventsResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(respBody, &body))
				assert.Equal(t, constant.SUCCESS.Message, body.Message)
				assert.Equal(t, 1, body.Data.Accepted)
				assert.Equal(t, 0, body.Data.Rejected)
				assert.Equal(t, 0, body.Data.Duplicate)

				// verify DB side effects
				var orders []models.Order
				db.Find(&orders)
				require.Len(t, orders, 1)
				assert.Equal(t, models.ORDER_STATUS_PAID, orders[0].CurrentStatus)

				var eventCount int64
				db.Model(&models.OrderEvent{}).Where("order_id = ?", orders[0].ID).Count(&eventCount)
				assert.Equal(t, int64(1), eventCount)
			},
		},
		{
			name: "invalid transition rejected",
			seedDB: func(t *testing.T) []dto.ImportOrderEventRequest {
				order := seedOrder(t, 5000, models.ORDER_STATUS_CREATED)
				return []dto.ImportOrderEventRequest{
					{OrderID: order.ID, Status: "delivered", EventAt: time.Now(), UpdatedBy: "admin"},
				}
			},
			expectedStatus: 200,
			validate: func(t *testing.T, respBody []byte) {
				var body struct {
					Data dto.ImportOrderEventsResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(respBody, &body))
				assert.Equal(t, 0, body.Data.Accepted)
				assert.Equal(t, 1, body.Data.Rejected)
				assert.Contains(t, body.Data.Errors[0].Reason, "Invalid transition")

				// verify order unchanged
				var orders []models.Order
				db.Where("current_status = ?", models.ORDER_STATUS_CREATED).Find(&orders)
				assert.Len(t, orders, 1)
			},
		},
		{
			name: "duplicate status",
			seedDB: func(t *testing.T) []dto.ImportOrderEventRequest {
				order := seedOrder(t, 5000, models.ORDER_STATUS_PAID)
				return []dto.ImportOrderEventRequest{
					{OrderID: order.ID, Status: "paid", EventAt: time.Now(), UpdatedBy: "admin"},
				}
			},
			expectedStatus: 200,
			validate: func(t *testing.T, respBody []byte) {
				var body struct {
					Data dto.ImportOrderEventsResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(respBody, &body))
				assert.Equal(t, 0, body.Data.Accepted)
				assert.Equal(t, 1, body.Data.Duplicate)
				assert.Contains(t, body.Data.Errors[0].Reason, "already in status")
			},
		},
		{
			name: "order not found",
			seedDB: func(t *testing.T) []dto.ImportOrderEventRequest {
				return []dto.ImportOrderEventRequest{
					{OrderID: 99999, Status: "paid", EventAt: time.Now(), UpdatedBy: "admin"},
				}
			},
			expectedStatus: 200,
			validate: func(t *testing.T, respBody []byte) {
				var body struct {
					Data dto.ImportOrderEventsResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(respBody, &body))
				assert.Equal(t, 0, body.Data.Accepted)
				assert.Equal(t, 1, body.Data.Rejected)
				assert.Contains(t, body.Data.Errors[0].Reason, "Order not found")
			},
		},
		{
			name: "validation failure - multiple invalid fields",
			seedDB: func(t *testing.T) []dto.ImportOrderEventRequest {
				return []dto.ImportOrderEventRequest{
					{OrderID: -1, Status: "paid", EventAt: time.Now(), UpdatedBy: "admin"},
					{OrderID: 1, Status: "unknown_status", EventAt: time.Now(), UpdatedBy: "admin"},
					{OrderID: 1, Status: "paid", UpdatedBy: "admin"}, // missing EventAt
				}
			},
			expectedStatus: 200,
			validate: func(t *testing.T, respBody []byte) {
				var body struct {
					Data dto.ImportOrderEventsResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(respBody, &body))
				assert.Equal(t, 0, body.Data.Accepted)
				assert.Equal(t, 3, body.Data.Rejected)
				assert.Len(t, body.Data.Errors, 3)
			},
		},
		{
			name:           "invalid JSON payload",
			rawBody:        []byte("{invalid-json"),
			expectedStatus: 400,
			validate: func(t *testing.T, respBody []byte) {
				var body struct {
					Status  int    `json:"status"`
					Message string `json:"message"`
				}
				require.NoError(t, json.Unmarshal(respBody, &body))
				assert.Equal(t, constant.INVALID_INPUT.Message, body.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanAll()

			var bodyBytes []byte
			if tc.rawBody != nil {
				bodyBytes = tc.rawBody
			} else {
				reqs := tc.seedDB(t)
				bodyBytes, _ = json.Marshal(reqs)
			}

			req := httptest.NewRequest("POST", "/api/v1/order-events/import", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", adminAPIKey)

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.validate != nil {
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				resp.Body.Close()
				tc.validate(t, respBody)
			}
		})
	}
}

func TestIntegrationImportOrderEventsFullLifecycle(t *testing.T) {
	cleanAll()

	order := seedOrder(t, 10000, models.ORDER_STATUS_CREATED)

	transitions := []string{"paid", "packed", "shipped", "delivered"}

	for _, status := range transitions {
		reqBody := []dto.ImportOrderEventRequest{
			{OrderID: order.ID, Status: status, EventAt: time.Now(), UpdatedBy: "admin"},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/order-events/import", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", adminAPIKey)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var body struct {
			Data dto.ImportOrderEventsResponse `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

		assert.Equal(t, 1, body.Data.Accepted, "expected accepted for transition to %s", status)
	}

	var finalOrder models.Order
	db.First(&finalOrder, order.ID)
	assert.Equal(t, models.ORDER_STATUS_DELIVERED, finalOrder.CurrentStatus)

	var eventCount int64
	db.Model(&models.OrderEvent{}).Where("order_id = ?", order.ID).Count(&eventCount)
	assert.Equal(t, int64(4), eventCount)
}

func TestIntegrationImportOrderEventsMixedBatch(t *testing.T) {
	cleanAll()

	order1 := seedOrder(t, 1000, models.ORDER_STATUS_CREATED)
	order2 := seedOrder(t, 2000, models.ORDER_STATUS_PAID)
	order3 := seedOrder(t, 3000, models.ORDER_STATUS_CREATED)

	reqBody := []dto.ImportOrderEventRequest{
		{OrderID: order1.ID, Status: "paid", EventAt: time.Now(), UpdatedBy: "admin"},
		{OrderID: order2.ID, Status: "paid", EventAt: time.Now(), UpdatedBy: "admin"},
		{OrderID: order3.ID, Status: "delivered", EventAt: time.Now(), UpdatedBy: "admin"},
		{OrderID: -1, Status: "paid", EventAt: time.Now(), UpdatedBy: "admin"},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/order-events/import", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", adminAPIKey)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body struct {
		Data dto.ImportOrderEventsResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, 1, body.Data.Accepted)
	assert.Equal(t, 2, body.Data.Rejected)
	assert.Equal(t, 1, body.Data.Duplicate)
}
