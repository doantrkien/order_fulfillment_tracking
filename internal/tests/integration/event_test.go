package integration

import (
	"bytes"
	"encoding/json"
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

func TestIntegrationImportOrderEventsSuccess(t *testing.T) {
	cleanAll()

	order := seedOrder(t, 5000, models.ORDER_STATUS_CREATED)

	reqBody := []dto.ImportOrderEventRequest{
		{OrderID: order.ID, Status: "paid", EventAt: time.Now(), UpdatedBy: "admin"},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/order-events/import", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body struct {
		Status  int                          `json:"status"`
		Message string                       `json:"message"`
		Data    dto.ImportOrderEventsResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, "batch processed", body.Message)
	assert.Equal(t, 1, body.Data.Accepted)
	assert.Equal(t, 0, body.Data.Rejected)
	assert.Equal(t, 0, body.Data.Duplicate)

	var updatedOrder models.Order
	db.First(&updatedOrder, order.ID)
	assert.Equal(t, models.ORDER_STATUS_PAID, updatedOrder.CurrentStatus)

	var eventCount int64
	db.Model(&models.OrderEvent{}).Where("order_id = ?", order.ID).Count(&eventCount)
	assert.Equal(t, int64(1), eventCount)
}

func TestIntegrationImportOrderEventsInvalidTransition(t *testing.T) {
	cleanAll()

	order := seedOrder(t, 5000, models.ORDER_STATUS_CREATED)

	reqBody := []dto.ImportOrderEventRequest{
		{OrderID: order.ID, Status: "delivered", EventAt: time.Now(), UpdatedBy: "admin"},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/order-events/import", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body struct {
		Data dto.ImportOrderEventsResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, 0, body.Data.Accepted)
	assert.Equal(t, 1, body.Data.Rejected)
	assert.Contains(t, body.Data.Errors[0].Reason, "Invalid transition")

	var unchangedOrder models.Order
	db.First(&unchangedOrder, order.ID)
	assert.Equal(t, models.ORDER_STATUS_CREATED, unchangedOrder.CurrentStatus)
}

func TestIntegrationImportOrderEventsDuplicateStatus(t *testing.T) {
	cleanAll()

	order := seedOrder(t, 5000, models.ORDER_STATUS_PAID)

	reqBody := []dto.ImportOrderEventRequest{
		{OrderID: order.ID, Status: "paid", EventAt: time.Now(), UpdatedBy: "admin"},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/order-events/import", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body struct {
		Data dto.ImportOrderEventsResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, 0, body.Data.Accepted)
	assert.Equal(t, 1, body.Data.Duplicate)
	assert.Contains(t, body.Data.Errors[0].Reason, "already in status")
}

func TestIntegrationImportOrderEventsOrderNotFound(t *testing.T) {
	cleanAll()

	reqBody := []dto.ImportOrderEventRequest{
		{OrderID: 99999, Status: "paid", EventAt: time.Now(), UpdatedBy: "admin"},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/order-events/import", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body struct {
		Data dto.ImportOrderEventsResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, 0, body.Data.Accepted)
	assert.Equal(t, 1, body.Data.Rejected)
	assert.Contains(t, body.Data.Errors[0].Reason, "Order not found")
}

func TestIntegrationImportOrderEventsValidationFailure(t *testing.T) {
	cleanAll()

	reqBody := []dto.ImportOrderEventRequest{
		{OrderID: -1, Status: "paid", EventAt: time.Now(), UpdatedBy: "admin"},
		{OrderID: 1, Status: "unknown_status", EventAt: time.Now(), UpdatedBy: "admin"},
		{OrderID: 1, Status: "paid", UpdatedBy: "admin"}, // missing EventAt
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/order-events/import", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body struct {
		Data dto.ImportOrderEventsResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, 0, body.Data.Accepted)
	assert.Equal(t, 3, body.Data.Rejected)
	assert.Len(t, body.Data.Errors, 3)
}

func TestIntegrationImportOrderEventsInvalidJSON(t *testing.T) {
	cleanAll()

	req := httptest.NewRequest("POST", "/api/v1/order-events/import", bytes.NewBuffer([]byte("{invalid-json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var body struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, "invalid payload", body.Message)
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
