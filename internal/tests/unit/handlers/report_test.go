package handlers_test

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"main/internal/handlers"
	"main/internal/models"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"

	"main/internal/tests/unit/mocks"
)

func setupReportHandlerTest(t *testing.T) (*fiber.App, *mocks.ReportService, *handlers.ReportHandler) {
	app := fiber.New()
	mockService := mocks.NewReportService(t)
	handler := handlers.NewReportHandler(mockService)

	return app, mockService, handler
}

func TestReportHandler_GetDailyReport(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		query          string
		setupMock      func(*mocks.ReportService)
		expectedStatus int
	}{
		{
			name:  "valid date",
			query: "date=2026-05-04",
			setupMock: func(mockService *mocks.ReportService) {
				requestDate := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
				mockService.On("GetDailyReport", requestDate).Return(&models.Report{ID: 1, Date: requestDate}, nil).Once()
			},
			expectedStatus: 200,
		},
		{
			name:           "invalid date format",
			query:          "date=2026-13-01",
			setupMock:      func(mockService *mocks.ReportService) {},
			expectedStatus: 400,
		},
		{
			name:  "repository error",
			query: "date=2026-05-04",
			setupMock: func(mockService *mocks.ReportService) {
				requestDate := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
				mockService.On("GetDailyReport", requestDate).Return((*models.Report)(nil), assert.AnError).Once()
			},
			expectedStatus: 500,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, mockService, handler := setupReportHandlerTest(t)
			app.Get("/reports", handler.GetDailyReport)
			tc.setupMock(mockService)

			req := httptest.NewRequest("GET", "/reports?"+tc.query, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestReportHandler_CreateDailyReport(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		body           string
		setupMock      func(*mocks.ReportService)
		expectedStatus int
	}{
		{
			name: "valid request",
			body: `{"date":"2026-05-04"}`,
			setupMock: func(mockService *mocks.ReportService) {
				requestDate := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
				mockService.On("CreateDailyReport", requestDate).Return(&models.Report{ID: 2, Date: requestDate}, nil).Once()
			},
			expectedStatus: 201,
		},
		{
			name:           "invalid json",
			body:           `{invalid-json}`,
			setupMock:      func(mockService *mocks.ReportService) {},
			expectedStatus: 400,
		},
		{
			name: "repository error",
			body: `{"date":"2026-05-04"}`,
			setupMock: func(mockService *mocks.ReportService) {
				requestDate := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
				mockService.On("CreateDailyReport", requestDate).Return((*models.Report)(nil), assert.AnError).Once()
			},
			expectedStatus: 500,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, mockService, handler := setupReportHandlerTest(t)
			app.Post("/reports", handler.CreateDailyReport)
			tc.setupMock(mockService)

			req := httptest.NewRequest("POST", "/reports", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}
