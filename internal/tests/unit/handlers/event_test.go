package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"main/constant"
	"main/internal/dto"
	"main/internal/handlers"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"main/internal/tests/unit/mocks"
)

func setupEventHandlerTest(t *testing.T) (*fiber.App, *mocks.OrderEventService, *handlers.OrderEventHandler) {
	app := fiber.New()
	mockService := mocks.NewOrderEventService(t)
	handler := handlers.NewOrderEventHandler(mockService)

	return app, mockService, handler
}

func TestOrderEventHandlerImportOrderEvents(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name           string
		body           interface{}
		setupMock      func(*mocks.OrderEventService)
		expectedStatus int
		validate       func(t *testing.T, respBody []byte)
	}{
		{
			name: "all accepted",
			body: []dto.ImportOrderEventRequest{
				{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
				{OrderID: 2, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			},
			setupMock: func(mockService *mocks.OrderEventService) {
				mockService.On("ImportOrderEvents", mock.AnythingOfType("[]dto.ImportOrderEventRequest")).Return(
					dto.ImportOrderEventsResponse{
						Accepted:  2,
						Rejected:  0,
						Duplicate: 0,
						Errors:    []dto.EventError{},
					}, nil,
				).Once()
			},
			expectedStatus: 200,
			validate: func(t *testing.T, respBody []byte) {
				var resp struct {
					Status  int                           `json:"status"`
					Message string                        `json:"message"`
					Data    dto.ImportOrderEventsResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(respBody, &resp))
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, constant.SUCCESS.Message, resp.Message)
				assert.Equal(t, 2, resp.Data.Accepted)
			},
		},
		{
			name:           "invalid payload - bad JSON",
			body:           nil, // handled separately with raw bytes
			setupMock:      func(mockService *mocks.OrderEventService) {},
			expectedStatus: 400,
			validate: func(t *testing.T, respBody []byte) {
				var resp struct {
					Status  int    `json:"status"`
					Message string `json:"message"`
				}
				require.NoError(t, json.Unmarshal(respBody, &resp))
				assert.Equal(t, 400, resp.Status)
				assert.Equal(t, constant.INVALID_INPUT.Message, resp.Message)
			},
		},
		{
			name: "service returns error",
			body: []dto.ImportOrderEventRequest{
				{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			},
			setupMock: func(mockService *mocks.OrderEventService) {
				mockService.On("ImportOrderEvents", mock.AnythingOfType("[]dto.ImportOrderEventRequest")).Return(
					dto.ImportOrderEventsResponse{
						Rejected: 1,
						Errors: []dto.EventError{
							{OrderID: 1, Status: "paid", Reason: "database connection lost"},
						},
					}, assert.AnError,
				).Once()
			},
			expectedStatus: 500,
			validate: func(t *testing.T, respBody []byte) {
				var resp struct {
					Status  int                           `json:"status"`
					Message string                        `json:"message"`
					Data    dto.ImportOrderEventsResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(respBody, &resp))
				assert.Equal(t, 500, resp.Status)
				assert.Equal(t, 1, resp.Data.Rejected)
			},
		},
		{
			name: "mixed results - partial success",
			body: []dto.ImportOrderEventRequest{
				{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
				{OrderID: 2, Status: "paid", EventAt: now, UpdatedBy: "admin"},
				{OrderID: 3, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			},
			setupMock: func(mockService *mocks.OrderEventService) {
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
			},
			expectedStatus: 200,
			validate: func(t *testing.T, respBody []byte) {
				var resp struct {
					Data dto.ImportOrderEventsResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(respBody, &resp))
				assert.Equal(t, 1, resp.Data.Accepted)
				assert.Equal(t, 1, resp.Data.Rejected)
				assert.Equal(t, 1, resp.Data.Duplicate)
				assert.Len(t, resp.Data.Errors, 2)
			},
		},
		{
			name: "empty request body",
			body: []dto.ImportOrderEventRequest{},
			setupMock: func(mockService *mocks.OrderEventService) {
				mockService.On("ImportOrderEvents", mock.AnythingOfType("[]dto.ImportOrderEventRequest")).Return(
					dto.ImportOrderEventsResponse{
						Accepted:  0,
						Rejected:  0,
						Duplicate: 0,
						Errors:    []dto.EventError{},
					}, nil,
				).Once()
			},
			expectedStatus: 200,
			validate: func(t *testing.T, respBody []byte) {
				var resp struct {
					Data dto.ImportOrderEventsResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(respBody, &resp))
				assert.Equal(t, 0, resp.Data.Accepted)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, mockService, handler := setupEventHandlerTest(t)
			app.Post("/events", handler.ImportOrderEvents)
			tc.setupMock(mockService)

			var reqBody []byte
			if tc.body == nil {
				// bad JSON case
				reqBody = []byte("{invalid-json}")
			} else {
				reqBody, _ = json.Marshal(tc.body)
			}

			req := httptest.NewRequest("POST", "/events", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)

			assert.NoError(t, err)
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
