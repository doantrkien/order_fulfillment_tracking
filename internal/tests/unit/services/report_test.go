package services_test

import (
	"testing"
	"time"

	"main/internal/models"
	"main/internal/services"
	"main/internal/tests/unit/mocks"

	"github.com/stretchr/testify/assert"
)

func TestReportService_GetDailyReport(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		requestDate    time.Time
		setupMock      func(*mocks.ReportRepository)
		expectError    bool
		expectedReport *models.Report
	}{
		{
			name:        "successful fetch",
			requestDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
			setupMock: func(mockRepo *mocks.ReportRepository) {
				expectedReport := &models.Report{ID: 10, Date: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)}
				mockRepo.On("GetDailyReport", time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)).Return(expectedReport, nil).Once()
			},
			expectedReport: &models.Report{ID: 10, Date: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)},
		},
		{
			name:        "repository error",
			requestDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
			setupMock: func(mockRepo *mocks.ReportRepository) {
				mockRepo.On("GetDailyReport", time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)).Return((*models.Report)(nil), assert.AnError).Once()
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewReportRepository(t)
			service := services.NewReportService(mockRepo)
			tc.setupMock(mockRepo)

			report, err := service.GetDailyReport(tc.requestDate)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, report)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedReport, report)
		})
	}
}

func TestReportService_CreateDailyReport(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		requestDate    time.Time
		setupMock      func(*mocks.ReportRepository)
		expectError    bool
		expectedReport *models.Report
	}{
		{
			name:        "successfully builds and saves report",
			requestDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
			setupMock: func(mockRepo *mocks.ReportRepository) {
				periodStart := time.Date(2026, 5, 4, 3, 0, 0, 0, time.UTC)
				periodEnd := periodStart.Add(24 * time.Hour)
				reportFromBuild := &models.Report{Date: periodStart, TotalOrders: 1}
				mockRepo.On("BuildDailyReport", periodStart, periodEnd).Return(reportFromBuild, nil).Once()
				mockRepo.On("SaveReport", reportFromBuild).Return(reportFromBuild, nil).Once()
			},
			expectedReport: &models.Report{Date: time.Date(2026, 5, 4, 3, 0, 0, 0, time.UTC), TotalOrders: 1},
		},
		{
			name:        "build report failure",
			requestDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
			setupMock: func(mockRepo *mocks.ReportRepository) {
				periodStart := time.Date(2026, 5, 4, 3, 0, 0, 0, time.UTC)
				periodEnd := periodStart.Add(24 * time.Hour)
				mockRepo.On("BuildDailyReport", periodStart, periodEnd).Return((*models.Report)(nil), assert.AnError).Once()
			},
			expectError: true,
		},
		{
			name:        "save report failure",
			requestDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
			setupMock: func(mockRepo *mocks.ReportRepository) {
				periodStart := time.Date(2026, 5, 4, 3, 0, 0, 0, time.UTC)
				periodEnd := periodStart.Add(24 * time.Hour)
				reportFromBuild := &models.Report{Date: periodStart, TotalOrders: 1}
				mockRepo.On("BuildDailyReport", periodStart, periodEnd).Return(reportFromBuild, nil).Once()
				mockRepo.On("SaveReport", reportFromBuild).Return((*models.Report)(nil), assert.AnError).Once()
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewReportRepository(t)
			service := services.NewReportService(mockRepo)
			tc.setupMock(mockRepo)

			report, err := service.CreateDailyReport(tc.requestDate)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, report)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedReport, report)
		})
	}
}
