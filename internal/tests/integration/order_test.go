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

func TestIntegration_CreateOrder_Success(t *testing.T) {
	cleanOrders()

	reqBody := dto.OrderRequest{
		TotalAmount:     5000,
		Username:        "integration_user",
		UserPhone:       "0901234567",
		ShippingAddress: "123 Test Street, HCM City",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-customer-key")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var order models.Order
	err = db.First(&order).Error
	require.NoError(t, err)
	assert.Equal(t, reqBody.TotalAmount, order.TotalAmount)
	assert.Equal(t, models.ORDER_STATUS_CREATED, order.CurrentStatus)

	var userInfo models.UserInfo
	json.Unmarshal(order.UserInfo, &userInfo)
	assert.Equal(t, reqBody.Username, userInfo.Username)
	assert.Equal(t, reqBody.UserPhone, userInfo.UserPhone)
	assert.Equal(t, reqBody.ShippingAddress, userInfo.ShippingAddress)
}

func TestIntegration_CreateOrder_InvalidJSON(t *testing.T) {
	cleanOrders()

	req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer([]byte("{invalid-json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-customer-key")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var count int64
	db.Model(&models.Order{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestIntegration_CreateOrder_Unauthenticated(t *testing.T) {
	cleanOrders()

	reqBody := dto.OrderRequest{TotalAmount: 1000}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestIntegration_CreateOrder_WrongRole(t *testing.T) {
	cleanOrders()

	reqBody := dto.OrderRequest{TotalAmount: 1000}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-admin-key")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestIntegration_GetAllOrders_Success(t *testing.T) {
	cleanOrders()

	seedOrders := []models.Order{
		{
			TotalAmount:   1000,
			CurrentStatus: models.ORDER_STATUS_CREATED,
			UserInfo:      mustMarshalUserInfo("alice", "0900000001", "Addr A"),
			CreatedAt:     time.Now().Add(-3 * time.Minute),
			UpdatedAt:     time.Now(),
		},
		{
			TotalAmount:   2000,
			CurrentStatus: models.ORDER_STATUS_PAID,
			UserInfo:      mustMarshalUserInfo("bob", "0900000002", "Addr B"),
			CreatedAt:     time.Now().Add(-2 * time.Minute),
			UpdatedAt:     time.Now(),
		},
		{
			TotalAmount:   3000,
			CurrentStatus: models.ORDER_STATUS_DELIVERED,
			UserInfo:      mustMarshalUserInfo("carol", "0900000003", "Addr C"),
			CreatedAt:     time.Now().Add(-1 * time.Minute),
			UpdatedAt:     time.Now(),
		},
	}
	db.Create(&seedOrders)

	req := httptest.NewRequest("GET", "/api/v1/orders?page=1&limit=10", nil)
	req.Header.Set("X-API-Key", "test-admin-key")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body struct {
		Status     int                `json:"status"`
		Data       []dto.OrderReponse `json:"data"`
		Pagination struct {
			Page       int   `json:"current_page"`
			Limit      int   `json:"limit_item"`
			TotalItems int64 `json:"total_items"`
			TotalPages int   `json:"total_pages"`
		} `json:"pagination"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, 200, body.Status)
	assert.Equal(t, int64(3), body.Pagination.TotalItems)
	assert.Len(t, body.Data, 3)

	assert.Equal(t, "carol", body.Data[0].Username)
	assert.Equal(t, models.ORDER_STATUS_DELIVERED, body.Data[0].Status)
}

func TestIntegration_GetAllOrders_FilterByStatus(t *testing.T) {
	cleanOrders()

	db.Create([]models.Order{
		{TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, UserInfo: mustMarshalUserInfo("u1", "", ""), CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{TotalAmount: 2000, CurrentStatus: models.ORDER_STATUS_PAID, UserInfo: mustMarshalUserInfo("u2", "", ""), CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{TotalAmount: 3000, CurrentStatus: models.ORDER_STATUS_PAID, UserInfo: mustMarshalUserInfo("u3", "", ""), CreatedAt: time.Now(), UpdatedAt: time.Now()},
	})

	req := httptest.NewRequest("GET", "/api/v1/orders?status=paid&page=1&limit=10", nil)
	req.Header.Set("X-API-Key", "test-admin-key")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body struct {
		Data       []dto.OrderReponse `json:"data"`
		Pagination struct {
			TotalItems int64 `json:"total_items"`
		} `json:"pagination"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, int64(2), body.Pagination.TotalItems)
	assert.Len(t, body.Data, 2)
	for _, o := range body.Data {
		assert.Equal(t, models.ORDER_STATUS_PAID, o.Status)
	}
}

func TestIntegration_GetAllOrders_FilterByDate(t *testing.T) {
	cleanOrders()

	today := time.Now()
	yesterday := time.Now().AddDate(0, 0, -1)

	db.Create([]models.Order{
		{TotalAmount: 1000, CurrentStatus: models.ORDER_STATUS_CREATED, UserInfo: mustMarshalUserInfo("today_user", "", ""), CreatedAt: today, UpdatedAt: today},
		{TotalAmount: 2000, CurrentStatus: models.ORDER_STATUS_CREATED, UserInfo: mustMarshalUserInfo("yesterday_user", "", ""), CreatedAt: yesterday, UpdatedAt: yesterday},
	})

	dateStr := today.Format("2006-01-02")
	req := httptest.NewRequest("GET", "/api/v1/orders?date="+dateStr+"&page=1&limit=10", nil)
	req.Header.Set("X-API-Key", "test-admin-key")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body struct {
		Data       []dto.OrderReponse `json:"data"`
		Pagination struct {
			TotalItems int64 `json:"total_items"`
		} `json:"pagination"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, int64(1), body.Pagination.TotalItems)
	assert.Len(t, body.Data, 1)
	assert.Equal(t, "today_user", body.Data[0].Username)
}

func TestIntegration_GetAllOrders_Pagination(t *testing.T) {
	cleanOrders()

	for i := 1; i <= 5; i++ {
		db.Create(&models.Order{
			TotalAmount:   int64(i * 1000),
			CurrentStatus: models.ORDER_STATUS_CREATED,
			UserInfo:      mustMarshalUserInfo("user", "", ""),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
	}

	req := httptest.NewRequest("GET", "/api/v1/orders?page=1&limit=2", nil)
	req.Header.Set("X-API-Key", "test-admin-key")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var body struct {
		Data       []dto.OrderReponse `json:"data"`
		Pagination struct {
			TotalItems int64 `json:"total_items"`
			TotalPages int   `json:"total_pages"`
			Page       int   `json:"current_page"`
			Limit      int   `json:"limit_item"`
		} `json:"pagination"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, int64(5), body.Pagination.TotalItems)
	assert.Equal(t, 3, body.Pagination.TotalPages)
	assert.Equal(t, 1, body.Pagination.Page)
	assert.Equal(t, 2, body.Pagination.Limit)
	assert.Len(t, body.Data, 2)
}

func TestIntegration_GetAllOrders_Unauthenticated(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/orders", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
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
