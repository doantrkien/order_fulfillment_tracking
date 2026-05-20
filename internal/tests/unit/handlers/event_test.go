package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"main/internal/dto"
	"main/internal/handlers"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"main/internal/tests/unit/mocks"
)

func setupEventHandlerTest(t *testing.T) (*fiber.App, *mocks.OrderEventService, *handlers.OrderEventHandler) {
	app := fiber.New()
	mockService := mocks.NewOrderEventService(t)
	handler := handlers.NewOrderEventHandler(mockService)

	return app, mockService, handler
}

func TestOrderEventHandlerImportOrderEvents(t *testing.T) {
	now := time.Now()

	t.Run("Success - all accepted", func(t *testing.T) {
		app, mockService, handler := setupEventHandlerTest(t)
		app.Post("/events", handler.ImportOrderEvents)

		reqBody := []dto.ImportOrderEventRequest{
			{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			{OrderID: 2, Status: "paid", EventAt: now, UpdatedBy: "admin"},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockService.On("ImportOrderEvents", mock.AnythingOfType("[]dto.ImportOrderEventRequest")).Return(
			dto.ImportOrderEventsResponse{
				Accepted:  2,
				Rejected:  0,
				Duplicate: 0,
				Errors:    []dto.EventError{},
			}, nil,
		).Once()

		req := httptest.NewRequest("POST", "/events", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var respBody struct {
			Status  int                          `json:"status"`
			Message string                       `json:"message"`
			Data    dto.ImportOrderEventsResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&respBody)

		assert.Equal(t, 200, respBody.Status)
		assert.Equal(t, "batch processed", respBody.Message)
		assert.Equal(t, 2, respBody.Data.Accepted)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid payload - bad JSON", func(t *testing.T) {
		app, _, handler := setupEventHandlerTest(t)
		app.Post("/events", handler.ImportOrderEvents)

		req := httptest.NewRequest("POST", "/events", bytes.NewBuffer([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		var respBody struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&respBody)

		assert.Equal(t, 400, respBody.Status)
		assert.Equal(t, "invalid payload", respBody.Message)
	})

	t.Run("Service returns error", func(t *testing.T) {
		app, mockService, handler := setupEventHandlerTest(t)
		app.Post("/events", handler.ImportOrderEvents)

		reqBody := []dto.ImportOrderEventRequest{
			{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockService.On("ImportOrderEvents", mock.AnythingOfType("[]dto.ImportOrderEventRequest")).Return(
			dto.ImportOrderEventsResponse{
				Rejected: 1,
				Errors: []dto.EventError{
					{OrderID: 1, Status: "paid", Reason: "database connection lost"},
				},
			}, errors.New("database connection lost"),
		).Once()

		req := httptest.NewRequest("POST", "/events", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode)

		var respBody struct {
			Status  int                          `json:"status"`
			Message string                       `json:"message"`
			Data    dto.ImportOrderEventsResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&respBody)

		assert.Equal(t, 500, respBody.Status)
		assert.Equal(t, "database connection lost", respBody.Message)
		assert.Equal(t, 1, respBody.Data.Rejected)
		mockService.AssertExpectations(t)
	})

	t.Run("Mixed results - partial success", func(t *testing.T) {
		app, mockService, handler := setupEventHandlerTest(t)
		app.Post("/events", handler.ImportOrderEvents)

		reqBody := []dto.ImportOrderEventRequest{
			{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			{OrderID: 2, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			{OrderID: 3, Status: "paid", EventAt: now, UpdatedBy: "admin"},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockService.On("ImportOrderEvents", mock.AnythingOfType("[]dto.ImportOrderEventRequest")).Return(
			dto.ImportOrderEventsResponse{
				Accepted:  1,
				Rejected:  1,
				Duplicate: 1,
				Errors: []dto.EventError{
					{OrderID: 2, Status: "paid", Reason: "Invalid transition"},
					{OrderID: 3, Status: "paid", Reason: "already in status 'paid'"},
				},
			}, nil,
		).Once()

		req := httptest.NewRequest("POST", "/events", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var respBody struct {
			Status  int                          `json:"status"`
			Message string                       `json:"message"`
			Data    dto.ImportOrderEventsResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&respBody)

		assert.Equal(t, 1, respBody.Data.Accepted)
		assert.Equal(t, 1, respBody.Data.Rejected)
		assert.Equal(t, 1, respBody.Data.Duplicate)
		assert.Len(t, respBody.Data.Errors, 2)
		mockService.AssertExpectations(t)
	})

	t.Run("Empty request body", func(t *testing.T) {
		app, mockService, handler := setupEventHandlerTest(t)
		app.Post("/events", handler.ImportOrderEvents)

		reqBody := []dto.ImportOrderEventRequest{}
		bodyBytes, _ := json.Marshal(reqBody)

		mockService.On("ImportOrderEvents", mock.AnythingOfType("[]dto.ImportOrderEventRequest")).Return(
			dto.ImportOrderEventsResponse{
				Accepted:  0,
				Rejected:  0,
				Duplicate: 0,
				Errors:    []dto.EventError{},
			}, nil,
		).Once()

		req := httptest.NewRequest("POST", "/events", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var respBody struct {
			Data dto.ImportOrderEventsResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&respBody)

		assert.Equal(t, 0, respBody.Data.Accepted)
		mockService.AssertExpectations(t)
	})
}
