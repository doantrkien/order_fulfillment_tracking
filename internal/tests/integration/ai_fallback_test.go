package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"main/internal/dto"
	"main/internal/models"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetFakeAIAdapter() {
	if testFakeAIAdapter != nil {
		testFakeAIAdapter.SimulateAIDisabled = false
		testFakeAIAdapter.SimulateTimeout = false
		testFakeAIAdapter.SimulateInvalidResponse = false
		testFakeAIAdapter.SimulateLowConfidence = false
		testFakeAIAdapter.SimulateConnectionError = false
		testFakeAIAdapter.ExpectedOutput = dto.ExceptionOutput{}
	}
}

func seedOrderWithNote(t *testing.T, totalAmount int64, status models.OrderStatus, note string, eventAt time.Time) models.Order {
	order := seedOrder(t, totalAmount, status)
	err := db.Create(&models.OrderEvent{
		OrderID:        order.ID,
		PreviousStatus: "",
		NewStatus:      status,
		EventAt:        eventAt,
		DriverNote:     &note,
	}).Error
	require.NoError(t, err)
	return order
}

func TestIntegrationAIFallbackFlow(t *testing.T) {
	t.Cleanup(func() {
		resetFakeAIAdapter()
		cleanAll()
	})

	t.Run("Happy Path - AI Success", func(t *testing.T) {
		resetFakeAIAdapter()
		cleanAll()

		// Seed a fresh order
		order := seedOrderWithNote(t, 10000, models.ORDER_STATUS_CREATED, "test", time.Now())

		// Mock standard successful AI response (must use a valid exception type, e.g. STUCK_ORDER)
		testFakeAIAdapter.ExpectedOutput = dto.ExceptionOutput{
			ExceptionType:      "STUCK_ORDER",
			Severity:           "MEDIUM",
			LikelyReason:       "AI thinks delivery is delayed due to weather",
			InternalNextAction: "Check weather report and call driver",
			Suggestion:         "Check weather report and call driver",
			ShouldAlert:        false,
			ConfidenceScore:    0.92,
		}

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/ai/orders/%d/exception-analysis", order.ID), bytes.NewBufferString(`{"note": "test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		var apiResp struct {
			Status  string                       `json:"status"`
			Message string                       `json:"message"`
			Data    dto.AnalyzeExceptionResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(respBody, &apiResp), "Response was: %s", string(respBody))

		// assert.Equal(t, 200, apiResp.Status)
		assert.Equal(t, "SUCCESS", apiResp.Status)
		assert.Equal(t, false, apiResp.Data.FallbackUsed)
		assert.Equal(t, "STUCK_ORDER", apiResp.Data.ExceptionType)
		assert.Equal(t, "MEDIUM", apiResp.Data.Severity)
		assert.Equal(t, "AI thinks delivery is delayed due to weather", apiResp.Data.LikelyReason)
		assert.Equal(t, "Check weather report and call driver", apiResp.Data.InternalNextAction)
		assert.InDelta(t, 0.92, apiResp.Data.ConfidenceScore, 0.001)
	})

	t.Run("Fallback Path - AI Connection Error", func(t *testing.T) {
		resetFakeAIAdapter()
		cleanAll()

		thirtyHoursAgo := time.Now().Add(-55 * time.Hour)
		// Seed order that is stuck (packed and updated 30 hours ago) to trigger STUCK_ORDER fallback with HIGH severity (base severity)
		order := seedOrderWithNote(t, 15000, models.ORDER_STATUS_CREATED, "test", thirtyHoursAgo)
		err := db.Model(&order).Updates(map[string]interface{}{
			"created_at": thirtyHoursAgo,
			"updated_at": thirtyHoursAgo,
		}).Error
		require.NoError(t, err)

		// Setup connection error simulation
		testFakeAIAdapter.SimulateConnectionError = true

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/ai/orders/%d/exception-analysis", order.ID), bytes.NewBufferString(`{"note": "test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode) // Assert API still returns 200 OK

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		var apiResp struct {
			Status  string                       `json:"status"`
			Message string                       `json:"message"`
			Data    dto.AnalyzeExceptionResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(respBody, &apiResp))

		// assert.Equal(t, 200, apiResp.Status)
		assert.Equal(t, "SUCCESS", apiResp.Status)
		assert.True(t, apiResp.Data.FallbackUsed)
		assert.Equal(t, "STUCK_ORDER", apiResp.Data.ExceptionType)
		assert.Equal(t, "HIGH", apiResp.Data.Severity)
	})

	t.Run("Fallback Path - AI Timeout", func(t *testing.T) {
		resetFakeAIAdapter()
		cleanAll()

		thirtyHoursAgo := time.Now().Add(-55 * time.Hour)
		order := seedOrderWithNote(t, 15000, models.ORDER_STATUS_CREATED, "test", thirtyHoursAgo)
		err := db.Model(&order).Updates(map[string]interface{}{
			"created_at": thirtyHoursAgo,
			"updated_at": thirtyHoursAgo,
		}).Error
		require.NoError(t, err)

		// Setup timeout simulation
		testFakeAIAdapter.SimulateTimeout = true

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/ai/orders/%d/exception-analysis", order.ID), bytes.NewBufferString(`{"note": "test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		var apiResp struct {
			Status  string                       `json:"status"`
			Message string                       `json:"message"`
			Data    dto.AnalyzeExceptionResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(respBody, &apiResp))

		// assert.Equal(t, 200, apiResp.Status)
		assert.Equal(t, "SUCCESS", apiResp.Status)
		assert.True(t, apiResp.Data.FallbackUsed)
		assert.Equal(t, "STUCK_ORDER", apiResp.Data.ExceptionType)
		assert.Equal(t, "HIGH", apiResp.Data.Severity)
	})

	t.Run("Fallback Path - AI Invalid JSON Response", func(t *testing.T) {
		resetFakeAIAdapter()
		cleanAll()

		thirtyHoursAgo := time.Now().Add(-55 * time.Hour)
		order := seedOrderWithNote(t, 15000, models.ORDER_STATUS_CREATED, "test", thirtyHoursAgo)
		err := db.Model(&order).Updates(map[string]interface{}{
			"created_at": thirtyHoursAgo,
			"updated_at": thirtyHoursAgo,
		}).Error
		require.NoError(t, err)

		// Setup invalid JSON simulation
		testFakeAIAdapter.SimulateInvalidResponse = true

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/ai/orders/%d/exception-analysis", order.ID), bytes.NewBufferString(`{"note": "test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		var apiResp struct {
			Status  string                       `json:"status"`
			Message string                       `json:"message"`
			Data    dto.AnalyzeExceptionResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(respBody, &apiResp))

		// assert.Equal(t, 200, apiResp.Status)
		assert.Equal(t, "SUCCESS", apiResp.Status)
		assert.True(t, apiResp.Data.FallbackUsed)
		assert.Equal(t, "STUCK_ORDER", apiResp.Data.ExceptionType)
		assert.Equal(t, "HIGH", apiResp.Data.Severity)
	})

	t.Run("Fallback Path - AI Low Confidence Score", func(t *testing.T) {
		resetFakeAIAdapter()
		cleanAll()

		thirtyHoursAgo := time.Now().Add(-55 * time.Hour)
		order := seedOrderWithNote(t, 15000, models.ORDER_STATUS_CREATED, "test", thirtyHoursAgo)
		err := db.Model(&order).Updates(map[string]interface{}{
			"created_at": thirtyHoursAgo,
			"updated_at": thirtyHoursAgo,
		}).Error
		require.NoError(t, err)

		// Setup low confidence simulation (confidence is 0.3, which is below the 0.5 threshold)
		testFakeAIAdapter.SimulateLowConfidence = true

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/ai/orders/%d/exception-analysis", order.ID), bytes.NewBufferString(`{"note": "test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		var apiResp struct {
			Status  string                       `json:"status"`
			Message string                       `json:"message"`
			Data    dto.AnalyzeExceptionResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(respBody, &apiResp))

		// assert.Equal(t, 200, apiResp.Status)
		assert.Equal(t, "SUCCESS", apiResp.Status)
		assert.True(t, apiResp.Data.FallbackUsed)
		assert.Equal(t, "STUCK_ORDER", apiResp.Data.ExceptionType)
		assert.Equal(t, "HIGH", apiResp.Data.Severity)
	})

	t.Run("Fallback Path - AI Disabled Via Adapter", func(t *testing.T) {
		resetFakeAIAdapter()
		cleanAll()

		thirtyHoursAgo := time.Now().Add(-55 * time.Hour)
		order := seedOrderWithNote(t, 15000, models.ORDER_STATUS_CREATED, "test", thirtyHoursAgo)
		err := db.Model(&order).Updates(map[string]interface{}{
			"created_at": thirtyHoursAgo,
			"updated_at": thirtyHoursAgo,
		}).Error
		require.NoError(t, err)

		// Setup AI Disabled simulation on the fake adapter
		testFakeAIAdapter.SimulateAIDisabled = true

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/ai/orders/%d/exception-analysis", order.ID), bytes.NewBufferString(`{"note": "test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		var apiResp struct {
			Status  string                       `json:"status"`
			Message string                       `json:"message"`
			Data    dto.AnalyzeExceptionResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(respBody, &apiResp))

		// assert.Equal(t, 200, apiResp.Status)
		assert.Equal(t, "SUCCESS", apiResp.Status)
		assert.True(t, apiResp.Data.FallbackUsed)
		assert.Equal(t, "STUCK_ORDER", apiResp.Data.ExceptionType)
		assert.Equal(t, "HIGH", apiResp.Data.Severity)
	})

	t.Run("Error Path - Order Not Found", func(t *testing.T) {
		resetFakeAIAdapter()
		cleanAll()

		req := httptest.NewRequest("POST", "/api/v1/ai/orders/999999/exception-analysis", bytes.NewBufferString(`{"note": "test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode) // Order doesn't exist, expect 404
	})
}
