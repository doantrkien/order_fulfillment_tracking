package services_test

import (
	"errors"
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

func TestOrderEventService_ImportOrderEvents(t *testing.T) {
	now := time.Now()

	t.Run("All events accepted", func(t *testing.T) {
		mockRepo, service := setupEventServiceTest(t, 2)

		reqs := []dto.ImportOrderEventRequest{
			{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			{OrderID: 2, Status: "paid", EventAt: now, UpdatedBy: "admin"},
		}

		mockRepo.On("ProcessSingleEventTx", mock.Anything).Return(
			repositories.ProcessResultDetail{Result: repositories.Accepted}, nil,
		).Times(2)

		resp, err := service.ImportOrderEvents(reqs)

		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Accepted)
		assert.Equal(t, 0, resp.Rejected)
		assert.Equal(t, 0, resp.Duplicate)
		assert.Empty(t, resp.Errors)
	})

	t.Run("Invalid order_id rejected by validation", func(t *testing.T) {
		_, service := setupEventServiceTest(t, 2)

		reqs := []dto.ImportOrderEventRequest{
			{OrderID: 0, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			{OrderID: -1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
		}

		resp, err := service.ImportOrderEvents(reqs)

		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Accepted)
		assert.Equal(t, 2, resp.Rejected)
		assert.Len(t, resp.Errors, 2)
		assert.Equal(t, "order_id must be greater than 0", resp.Errors[0].Reason)
	})

	t.Run("Invalid status rejected by validation", func(t *testing.T) {
		_, service := setupEventServiceTest(t, 2)

		reqs := []dto.ImportOrderEventRequest{
			{OrderID: 1, Status: "unknown_status", EventAt: now, UpdatedBy: "admin"},
		}

		resp, err := service.ImportOrderEvents(reqs)

		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Accepted)
		assert.Equal(t, 1, resp.Rejected)
		assert.Equal(t, "unknown status value", resp.Errors[0].Reason)
	})

	t.Run("Missing event_at rejected by validation", func(t *testing.T) {
		_, service := setupEventServiceTest(t, 2)

		reqs := []dto.ImportOrderEventRequest{
			{OrderID: 1, Status: "paid", UpdatedBy: "admin"}, // EventAt is zero
		}

		resp, err := service.ImportOrderEvents(reqs)

		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Accepted)
		assert.Equal(t, 1, resp.Rejected)
		assert.Equal(t, "event_at is required", resp.Errors[0].Reason)
	})

	t.Run("Duplicate events counted correctly", func(t *testing.T) {
		mockRepo, service := setupEventServiceTest(t, 2)

		reqs := []dto.ImportOrderEventRequest{
			{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
		}

		mockRepo.On("ProcessSingleEventTx", mock.Anything).Return(
			repositories.ProcessResultDetail{
				Result: repositories.Duplicate,
				Reason: "Order is already in status 'paid'",
			}, nil,
		).Once()

		resp, err := service.ImportOrderEvents(reqs)

		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Accepted)
		assert.Equal(t, 1, resp.Duplicate)
		assert.Len(t, resp.Errors, 1)
		assert.Contains(t, resp.Errors[0].Reason, "already in status")
	})

	t.Run("Rejected by repo (invalid transition)", func(t *testing.T) {
		mockRepo, service := setupEventServiceTest(t, 2)

		reqs := []dto.ImportOrderEventRequest{
			{OrderID: 1, Status: "delivered", EventAt: now, UpdatedBy: "admin"},
		}

		mockRepo.On("ProcessSingleEventTx", mock.Anything).Return(
			repositories.ProcessResultDetail{
				Result: repositories.Rejected,
				Reason: "Invalid transition from 'created' to 'delivered'",
			}, nil,
		).Once()

		resp, err := service.ImportOrderEvents(reqs)

		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Accepted)
		assert.Equal(t, 1, resp.Rejected)
		assert.Contains(t, resp.Errors[0].Reason, "Invalid transition")
	})

	t.Run("Repository error propagated", func(t *testing.T) {
		mockRepo, service := setupEventServiceTest(t, 2)

		reqs := []dto.ImportOrderEventRequest{
			{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
		}

		mockRepo.On("ProcessSingleEventTx", mock.Anything).Return(
			repositories.ProcessResultDetail{}, errors.New("database connection lost"),
		).Once()

		resp, err := service.ImportOrderEvents(reqs)

		assert.Error(t, err)
		assert.Equal(t, 1, resp.Rejected)
		assert.Equal(t, "database connection lost", resp.Errors[0].Reason)
	})

	t.Run("Mixed results - validation fail + accepted + duplicate", func(t *testing.T) {
		mockRepo, service := setupEventServiceTest(t, 2)

		reqs := []dto.ImportOrderEventRequest{
			{OrderID: 0, Status: "paid", EventAt: now, UpdatedBy: "admin"}, // validation fail
			{OrderID: 1, Status: "paid", EventAt: now, UpdatedBy: "admin"}, // will be accepted
			{OrderID: 2, Status: "paid", EventAt: now, UpdatedBy: "admin"}, // will be duplicate
		}

		// Use a function matcher to return different results based on OrderID
		mockRepo.On("ProcessSingleEventTx", mock.MatchedBy(func(e models.OrderEvent) bool {
			return e.OrderID == 1
		})).Return(
			repositories.ProcessResultDetail{Result: repositories.Accepted}, nil,
		).Once()

		mockRepo.On("ProcessSingleEventTx", mock.MatchedBy(func(e models.OrderEvent) bool {
			return e.OrderID == 2
		})).Return(
			repositories.ProcessResultDetail{
				Result: repositories.Duplicate,
				Reason: "Order is already in status 'paid'",
			}, nil,
		).Once()

		resp, err := service.ImportOrderEvents(reqs)

		assert.NoError(t, err)
		assert.Equal(t, 1, resp.Accepted)
		assert.Equal(t, 1, resp.Rejected) // validation fail
		assert.Equal(t, 1, resp.Duplicate)
	})

	t.Run("Empty request list", func(t *testing.T) {
		_, service := setupEventServiceTest(t, 2)

		resp, err := service.ImportOrderEvents([]dto.ImportOrderEventRequest{})

		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Accepted)
		assert.Equal(t, 0, resp.Rejected)
		assert.Equal(t, 0, resp.Duplicate)
	})

	t.Run("All events fail validation - repo never called", func(t *testing.T) {
		_, service := setupEventServiceTest(t, 2)

		reqs := []dto.ImportOrderEventRequest{
			{OrderID: -1, Status: "paid", EventAt: now, UpdatedBy: "admin"},
			{OrderID: 1, Status: "invalid_status", EventAt: now, UpdatedBy: "admin"},
		}

		resp, err := service.ImportOrderEvents(reqs)

		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Accepted)
		assert.Equal(t, 2, resp.Rejected)
		// The mock's Cleanup will verify ProcessSingleEventTx was NEVER called
	})
}
