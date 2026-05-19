package handlers_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"net/http/httptest"

	"main/internal/dto"
	"main/internal/handlers"
	"main/internal/models"
	"main/internal/services/mocks"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func setupOrderHandlerTest(t *testing.T) (*fiber.App, *mocks.OrderService, *handlers.OrderHandler) {
	app := fiber.New()
	mockService := mocks.NewOrderService(t)
	handler := handlers.NewOrderHandler(mockService)

	return app, mockService, handler
}

func TestOrderHandler_CreateOrder(t *testing.T) {
	app, mockService, handler := setupOrderHandlerTest(t)
	app.Post("/orders", handler.CreateOrder)

	t.Run("Success", func(t *testing.T) {
		reqBody := dto.OrderRequest{
			TotalAmount:     1000,
			Username:        "testuser",
			UserPhone:       "0123456789",
			ShippingAddress: "Test Address",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockService.On("CreateOrder", reqBody).Return(&models.Order{ID: 1}, nil).Once()

		req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestOrderHandler_GetAllOrder(t *testing.T) {
	app, mockService, handler := setupOrderHandlerTest(t)
	app.Get("/orders", handler.GetAllOrder)

	t.Run("Success with default pagination", func(t *testing.T) {
		mockResponse := []dto.OrderReponse{
			{
				ID:          1,
				TotalAmount: 1000,
				Status:      models.ORDER_STATUS_CREATED,
				Ordered_at:  time.Now(),
			},
		}

		expectedQuery := dto.OrderQuery{
			Page:  1,
			Limit: 10,
		}

		mockService.On("GetAllOrder", expectedQuery).Return(mockResponse, int64(1), nil).Once()

		req := httptest.NewRequest("GET", "/orders", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var respBody struct {
			Status     int                `json:"status"`
			Data       []dto.OrderReponse `json:"data"`
			Pagination struct {
				Page  int `json:"current_page"`
				Limit int `json:"limit_item"`
				Total int `json:"total_items"`
			} `json:"pagination"`
		}
		json.NewDecoder(resp.Body).Decode(&respBody)

		assert.Equal(t, 1, respBody.Pagination.Total)
		mockService.AssertExpectations(t)
	})
}
