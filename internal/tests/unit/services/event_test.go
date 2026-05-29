package services_test

import (
	"context"
	"testing"
	"time"

	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"
	"main/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"main/internal/tests/unit/mocks"
)

func setupEventServiceTest(t *testing.T, maxWorkers int) (*mocks.OrderEventRepository, services.OrderEventService) {
	mockRepo := mocks.NewOrderEventRepository(t)
	service := services.NewOrderEventService(mockRepo, maxWorkers)
	return mockRepo, service
}

func TestOrderEventServiceImportOrderEvents(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name      string
		requests  []dto.ImportOrderEventRequest
		setupMock func(*mocks.OrderEventRepository)
		expectErr bool
		validate  func(t *testing.T, resp dto.ImportOrderEventsResponse)
	}{
		{
			name: "all events accepted",
			requests: []dto.ImportOrderEventRequest{
				{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
				{OrderID: 2, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			},
			setupMock: func(mockRepo *mocks.OrderEventRepository) {
				mockRepo.On("ProcessBatchEventsTx", mock.Anything, mock.Anything).Return(
					func(_ context.Context, events []models.OrderEvent) []repositories.ProcessResultDetail {
						details := make([]repositories.ProcessResultDetail, len(events))
						for i, e := range events {
							details[i] = repositories.ProcessResultDetail{
								Result:  repositories.Accepted,
								OrderID: e.OrderID,
								Status:  string(e.NewStatus),
							}
						}
						return details
					}, nil,
				)
			},
			validate: func(t *testing.T, resp dto.ImportOrderEventsResponse) {
				assert.Equal(t, 2, resp.Accepted)
				assert.Equal(t, 0, resp.Rejected)
				assert.Equal(t, 0, resp.Duplicate)
				assert.Empty(t, resp.Errors)
			},
		},
		{
			name: "invalid order_id rejected by validation",
			requests: []dto.ImportOrderEventRequest{
				{OrderID: 0, Status: "paid", EventAt: now, UpdatedBy: "admin"},
				{OrderID: -1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			},
			setupMock: func(mockRepo *mocks.OrderEventRepository) {},
			validate: func(t *testing.T, resp dto.ImportOrderEventsResponse) {
				assert.Equal(t, 0, resp.Accepted)
				assert.Equal(t, 2, resp.Rejected)
				assert.Len(t, resp.Errors, 2)
				assert.Equal(t, "order_id must be greater than 0", resp.Errors[0].Reason)
			},
		},
		{
			name: "invalid status rejected by validation",
			requests: []dto.ImportOrderEventRequest{
				{OrderID: 1, Status: "unknown_status", EventAt: now, UpdatedBy: "admin"},
			},
			setupMock: func(mockRepo *mocks.OrderEventRepository) {},
			validate: func(t *testing.T, resp dto.ImportOrderEventsResponse) {
				assert.Equal(t, 0, resp.Accepted)
				assert.Equal(t, 1, resp.Rejected)
				assert.Equal(t, "unknown status value", resp.Errors[0].Reason)
			},
		},
		{
			name: "missing event_at rejected by validation",
			requests: []dto.ImportOrderEventRequest{
				{OrderID: 1, Status: "paid", UpdatedBy: "admin"}, // EventAt is zero
			},
			setupMock: func(mockRepo *mocks.OrderEventRepository) {},
			validate: func(t *testing.T, resp dto.ImportOrderEventsResponse) {
				assert.Equal(t, 0, resp.Accepted)
				assert.Equal(t, 1, resp.Rejected)
				assert.Equal(t, "event_at is required", resp.Errors[0].Reason)
			},
		},
		{
			name: "duplicate events counted correctly",
			requests: []dto.ImportOrderEventRequest{
				{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			},
			setupMock: func(mockRepo *mocks.OrderEventRepository) {
				mockRepo.On("ProcessBatchEventsTx", mock.Anything, mock.Anything).Return(
					[]repositories.ProcessResultDetail{
						{
							Result:  repositories.Duplicate,
							Reason:  "Order is already in status 'paid'",
							OrderID: 1,
							Status:  "paid",
						},
					}, nil,
				)
			},
			validate: func(t *testing.T, resp dto.ImportOrderEventsResponse) {
				assert.Equal(t, 0, resp.Accepted)
				assert.Equal(t, 1, resp.Duplicate)
				assert.Len(t, resp.Errors, 1)
				assert.Contains(t, resp.Errors[0].Reason, "already in status")
			},
		},
		{
			name: "rejected by repo - invalid transition",
			requests: []dto.ImportOrderEventRequest{
				{OrderID: 1, Status: "delivered", EventAt: now, UpdatedBy: "admin"},
			},
			setupMock: func(mockRepo *mocks.OrderEventRepository) {
				mockRepo.On("ProcessBatchEventsTx", mock.Anything, mock.Anything).Return(
					[]repositories.ProcessResultDetail{
						{
							Result:  repositories.Rejected,
							Reason:  "Invalid transition from 'created' to 'delivered'",
							OrderID: 1,
							Status:  "delivered",
						},
					}, nil,
				)
			},
			validate: func(t *testing.T, resp dto.ImportOrderEventsResponse) {
				assert.Equal(t, 0, resp.Accepted)
				assert.Equal(t, 1, resp.Rejected)
				assert.Contains(t, resp.Errors[0].Reason, "Invalid transition")
			},
		},
		{
			name: "repository error propagated",
			requests: []dto.ImportOrderEventRequest{
				{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			},
			setupMock: func(mockRepo *mocks.OrderEventRepository) {
				mockRepo.On("ProcessBatchEventsTx", mock.Anything, mock.Anything).Return(
					[]repositories.ProcessResultDetail(nil), assert.AnError,
				)
			},
			expectErr: true,
			validate: func(t *testing.T, resp dto.ImportOrderEventsResponse) {
				// When batch fails entirely, no individual results are tallied
				assert.Equal(t, 0, resp.Accepted)
			},
		},
		{
			name: "mixed results - validation fail + accepted + duplicate",
			requests: []dto.ImportOrderEventRequest{
				{OrderID: 0, Status: "paid", EventAt: now, UpdatedBy: "admin"}, // validation fail
				{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"}, // will be accepted
				{OrderID: 2, Status: "paid", EventAt: now, UpdatedBy: "admin"}, // will be duplicate
			},
			setupMock: func(mockRepo *mocks.OrderEventRepository) {
				mockRepo.On("ProcessBatchEventsTx", mock.Anything, mock.Anything).Return(
					func(_ context.Context, events []models.OrderEvent) []repositories.ProcessResultDetail {
						details := make([]repositories.ProcessResultDetail, len(events))
						for i, e := range events {
							switch e.OrderID {
							case 1:
								details[i] = repositories.ProcessResultDetail{
									Result:  repositories.Accepted,
									OrderID: 1,
									Status:  "paid",
								}
							case 2:
								details[i] = repositories.ProcessResultDetail{
									Result:  repositories.Duplicate,
									Reason:  "Order is already in status 'paid'",
									OrderID: 2,
									Status:  "paid",
								}
							}
						}
						return details
					}, nil,
				)
			},
			validate: func(t *testing.T, resp dto.ImportOrderEventsResponse) {
				assert.Equal(t, 1, resp.Accepted)
				assert.Equal(t, 1, resp.Rejected) // validation fail
				assert.Equal(t, 1, resp.Duplicate)
			},
		},
		{
			name:      "empty request list",
			requests:  []dto.ImportOrderEventRequest{},
			setupMock: func(mockRepo *mocks.OrderEventRepository) {},
			validate: func(t *testing.T, resp dto.ImportOrderEventsResponse) {
				assert.Equal(t, 0, resp.Accepted)
				assert.Equal(t, 0, resp.Rejected)
				assert.Equal(t, 0, resp.Duplicate)
			},
		},
		{
			name: "all events fail validation - repo never called",
			requests: []dto.ImportOrderEventRequest{
				{OrderID: -1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
				{OrderID: 1, Status: "invalid_status", EventAt: now, UpdatedBy: "admin"},
			},
			setupMock: func(mockRepo *mocks.OrderEventRepository) {
				// The mock's Cleanup will verify ProcessBatchEventsTx was NEVER called
			},
			validate: func(t *testing.T, resp dto.ImportOrderEventsResponse) {
				assert.Equal(t, 0, resp.Accepted)
				assert.Equal(t, 2, resp.Rejected)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo, service := setupEventServiceTest(t, 2)
			tc.setupMock(mockRepo)

			resp, err := service.ImportOrderEvents(context.Background(), tc.requests)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.validate != nil {
				tc.validate(t, resp)
			}
		})
	}
}
