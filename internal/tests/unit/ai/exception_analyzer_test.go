package ai

import (
	"context"
	"errors"

	"main/constant"
	"main/internal/ai"
	dto_ai "main/internal/dto/ai"
	dto_api "main/internal/dto/api"
	"main/internal/models"
	"main/utils/helpers"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAdapter implements AIAdapter for testing.
type mockAdapter struct {
	output    dto_ai.AIAnalysisResult
	outputStr string
	err       error
}

func (m *mockAdapter) AnalyzeException(_ context.Context, _ dto_ai.ExceptionPromptContext) (dto_ai.AIAnalysisResult, string, error) {
	return m.output, m.outputStr, m.err
}

func (m *mockAdapter) SummarizeReport(_ context.Context, _ dto_ai.AIAnalysisResult) (dto_api.ReportSummaryOutput, error) {
	return dto_api.ReportSummaryOutput{}, nil
}

func (m *mockAdapter) Ping(_ context.Context) error {
	return nil
}

func (m *mockAdapter) DraftCustomerUpdate(_ context.Context, _ dto_ai.CustomerUpdateDraftInput) (string, error) {
	return m.outputStr, m.err
}

func (m *mockAdapter) ReloadKnowledge(_ context.Context) error {
	return nil
}

func newTestAIContext() *models.AIContext {
	now := time.Now()
	return &models.AIContext{
		OrderID:       1001,
		CreatedAt:     now.Add(-2 * time.Hour),
		CurrentStatus: models.ORDER_STATUS_PAID,
		TotalAmount:   500000,
		Events: []models.AIEvent{
			{
				EventAt:        now.Add(-1 * time.Hour),
				PreviousStatus: models.ORDER_STATUS_CREATED,
				NewStatus:      models.ORDER_STATUS_PAID,
				UpdatedBy:      "admin_1",
			},
		},
	}
}

func newTestAIContextWithDriverNote(note string) *models.AIContext {
	ctx := newTestAIContext()
	ctx.Events[0].DriverNote = &note
	return ctx
}

func TestExceptionAnalyzer_AIDisabled(t *testing.T) {
	analyzer := ai.NewExceptionAnalyzer(nil, ai.ExceptionAnalyzerConfig{
		AIEnabled: false,
		AITimeout: 10 * time.Second,
	})

	aiCtx := newTestAIContext()
	// result, err := analyzer.Analyze(context.Background(), aiCtx, "")
	result, err := analyzer.Analyze(context.Background(), aiCtx)
	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, constant.FallbackReasonDisabled, result.FallbackReason)
	assert.Equal(t, "", result.RawResponse)
}

// TestExceptionAnalyzer_NoDriverNote verifies that when AI is enabled but no event
// carries a driver note, AI is NOT called and rule-based result is returned directly.
func TestExceptionAnalyzer_NoDriverNote_SkipsAI(t *testing.T) {
	// adapter would panic if called — ensures AI is never invoked
	analyzer := ai.NewExceptionAnalyzer(nil, ai.ExceptionAnalyzerConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	aiCtx := newTestAIContext() // no driver notes
	// result, err := analyzer.Analyze(context.Background(), aiCtx, "")
	result, err := analyzer.Analyze(context.Background(), aiCtx)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, constant.FallbackReasonNoDriverNote, result.FallbackReason)
	assert.Equal(t, "", result.RawResponse)
}

// TestExceptionAnalyzer_WithDriverNote_CallsAI verifies that when a driver note
// is present, the AI adapter is invoked.
func TestExceptionAnalyzer_WithDriverNote_CallsAI(t *testing.T) {
	adapter := &mockAdapter{
		output: dto_ai.AIAnalysisResult{
			ExceptionType:      "DELIVERY_FAILURE",
			Severity:           "HIGH",
			LikelyReason:       "Driver note indicates vehicle breakdown",
			InternalNextAction: "Reschedule delivery",
			ConfidenceScore:    0.9,
		},
		outputStr: `{"exception_type":"DELIVERY_FAILURE"}`,
	}
	analyzer := ai.NewExceptionAnalyzer(adapter, ai.ExceptionAnalyzerConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	aiCtx := newTestAIContextWithDriverNote("vehicle breakdown on highway")
	// result, err := analyzer.Analyze(context.Background(), aiCtx, "")
	result, err := analyzer.Analyze(context.Background(), aiCtx)

	require.NoError(t, err)
	assert.False(t, result.FallbackUsed)
	assert.Equal(t, "DELIVERY_FAILURE", result.ExceptionType)
	assert.Equal(t, 0.9, result.ConfidenceScore)
}

func TestExceptionAnalyzer_AIReturnsError(t *testing.T) {
	adapter := &mockAdapter{
		err: errors.New("connection refused"),
	}
	analyzer := ai.NewExceptionAnalyzer(adapter, ai.ExceptionAnalyzerConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	aiCtx := newTestAIContextWithDriverNote("giao hàng thất bại vì hỏng xe giữa đường") // need actionable driver note to trigger AI
	// result, err := analyzer.Analyze(context.Background(), aiCtx, "")
	result, err := analyzer.Analyze(context.Background(), aiCtx)

	require.NoError(t, err) // Analyzer should NOT return error on AI failure
	assert.True(t, result.FallbackUsed)
	assert.Contains(t, []string{constant.FallbackReasonTimeout, constant.FallbackReasonConnectionError}, result.FallbackReason)
	assert.GreaterOrEqual(t, result.DurationMs, 0)
}

func TestExceptionAnalyzer_AIReturnsTimeout(t *testing.T) {
	adapter := &mockAdapter{
		err: context.DeadlineExceeded,
	}
	analyzer := ai.NewExceptionAnalyzer(adapter, ai.ExceptionAnalyzerConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	aiCtx := newTestAIContextWithDriverNote("giao hàng thất bại vì hỏng xe giữa đường") // need actionable driver note to trigger AI
	// result, err := analyzer.Analyze(context.Background(), aiCtx, "")
	result, err := analyzer.Analyze(context.Background(), aiCtx)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, constant.FallbackReasonTimeout, result.FallbackReason)
}

func TestExceptionAnalyzer_AIReturnsInvalidResponse(t *testing.T) {
	// The adapter returns an error for invalid responses
	adapter := &mockAdapter{
		outputStr: "{invalid-json}",
		err:       errors.New("AI response is not valid JSON"),
	}
	analyzer := ai.NewExceptionAnalyzer(adapter, ai.ExceptionAnalyzerConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	aiCtx := newTestAIContextWithDriverNote("giao hàng thất bại vì hỏng xe giữa đường") // need actionable driver note to trigger AI
	// result, err := analyzer.Analyze(context.Background(), aiCtx, "")
	result, err := analyzer.Analyze(context.Background(), aiCtx)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Contains(t, []string{constant.FallbackReasonConnectionError, constant.FallbackReasonTimeout, constant.FallbackReasonInvalidResponse}, result.FallbackReason)
	assert.Equal(t, "{invalid-json}", result.RawResponse)
}

func TestExceptionAnalyzer_AIReturnsLowConfidence(t *testing.T) {
	// Valid output but confidence below threshold (0.6)
	adapter := &mockAdapter{
		output: dto_ai.AIAnalysisResult{
			ExceptionType:      "STUCK_ORDER",
			Severity:           "HIGH",
			LikelyReason:       "Some reason that is valid",
			InternalNextAction: "Some action that is valid",
			ConfidenceScore:    0.3,
		},
		outputStr: `{"confidence_score": 0.3}`,
	}
	analyzer := ai.NewExceptionAnalyzer(adapter, ai.ExceptionAnalyzerConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	aiCtx := newTestAIContextWithDriverNote("giao hàng thất bại vì hỏng xe giữa đường") // need actionable driver note to trigger AI
	// result, err := analyzer.Analyze(context.Background(), aiCtx, "test notes")
	result, err := analyzer.Analyze(context.Background(), aiCtx)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, constant.FallbackReasonLowConfidence, result.FallbackReason)
	assert.Equal(t, `{"confidence_score": 0.3}`, result.RawResponse)
}

func TestExceptionAnalyzer_FallbackProducesValidResult(t *testing.T) {
	// AI disabled, but the order has an invalid transition in its events
	analyzer := ai.NewExceptionAnalyzer(nil, ai.ExceptionAnalyzerConfig{
		AIEnabled: false,
		AITimeout: 10 * time.Second,
	})

	now := time.Now()
	aiCtx := &models.AIContext{
		OrderID:       2001,
		CreatedAt:     now.Add(-5 * time.Hour),
		CurrentStatus: models.ORDER_STATUS_SHIPPED,
		TotalAmount:   100000,
		Events: []models.AIEvent{
			{
				EventAt:        now.Add(-4 * time.Hour),
				PreviousStatus: models.ORDER_STATUS_CREATED,
				NewStatus:      models.ORDER_STATUS_DELIVERED, // Invalid!
				UpdatedBy:      "admin_1",
			},
		},
	}

	// result, err := analyzer.Analyze(context.Background(), aiCtx, "")
	result, err := analyzer.Analyze(context.Background(), aiCtx)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, constant.FallbackReasonDisabled, result.FallbackReason)
	assert.Equal(t, "INVALID_TRANSITION", result.ExceptionType)
	assert.Equal(t, "CRITICAL", result.Severity)
	assert.Equal(t, float64(1), result.ConfidenceScore)
	assert.NotEmpty(t, result.LikelyReason)
	assert.NotEmpty(t, result.InternalNextAction)
}

func TestExceptionAnalyzer_FallbackNoExceptionDetected(t *testing.T) {
	// AI disabled, order is healthy (no rules triggered)
	analyzer := ai.NewExceptionAnalyzer(nil, ai.ExceptionAnalyzerConfig{
		AIEnabled: false,
		AITimeout: 10 * time.Second,
	})

	now := time.Now()
	aiCtx := &models.AIContext{
		OrderID:       3001,
		CreatedAt:     now.Add(-1 * time.Hour),
		CurrentStatus: models.ORDER_STATUS_PAID,
		Events: []models.AIEvent{
			{
				EventAt:        now.Add(-30 * time.Minute),
				PreviousStatus: models.ORDER_STATUS_CREATED,
				NewStatus:      models.ORDER_STATUS_PAID,
				UpdatedBy:      "admin_1",
			},
		},
	}

	// result, err := analyzer.Analyze(context.Background(), aiCtx, "")
	result, err := analyzer.Analyze(context.Background(), aiCtx)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, "OTHER", result.ExceptionType)
	assert.Equal(t, "LOW", result.Severity)
}

func TestExceptionAnalyzer_NeverReturnsError_ForAIFailures(t *testing.T) {
	// Verify the contract: Analyze() never returns error for AI-related failures
	testErrors := []error{
		errors.New("connection refused"),
		context.DeadlineExceeded,
		context.Canceled,
		errors.New("unexpected EOF"),
	}

	for _, testErr := range testErrors {
		adapter := &mockAdapter{err: testErr}
		analyzer := ai.NewExceptionAnalyzer(adapter, ai.ExceptionAnalyzerConfig{
			AIEnabled: true,
			AITimeout: 5 * time.Second,
		})

		// Must have a driver note so the AI path is exercised
		aiCtx := newTestAIContextWithDriverNote("giao hàng thất bại vì hỏng xe giữa đường")
		// result, err := analyzer.Analyze(context.Background(), aiCtx, "")
		result, err := analyzer.Analyze(context.Background(), aiCtx)
		assert.NoError(t, err, "Analyze should not return error for: %v", testErr)
		assert.True(t, result.FallbackUsed, "Fallback should be used for: %v", testErr)
	}
}

func TestBuildExceptionInput(t *testing.T) {
	now := time.Now()
	aiCtx := &models.AIContext{
		OrderID:         42,
		CurrentStatus:   models.ORDER_STATUS_SHIPPED,
		TotalAmount:     1000,
		CustomerName:    "Test",
		ShippingAddress: "123 St",
		CreatedAt:       now,
		Events: []models.AIEvent{
			{
				EventAt:        now,
				PreviousStatus: models.ORDER_STATUS_PACKED,
				NewStatus:      models.ORDER_STATUS_SHIPPED,
				UpdatedBy:      "driver_5",
				DriverNote:     func(s string) *string { return &s }("customer complained"),
			},
		},
	}

	// input := buildExceptionInput(aiCtx, "customer complained")
	input := helpers.BuildExceptionInput(aiCtx)

	assert.Equal(t, int64(42), input.OrderID)
	assert.Equal(t, "shipped", input.CurrentStatus)
	assert.Equal(t, int64(1000), input.TotalAmount)
	assert.Equal(t, "Test", input.CustomerName)
	assert.Equal(t, "123 St", input.ShippingAddress)
	assert.Equal(t, now.Format(time.RFC3339), input.CreatedAt)
	assert.Equal(t, "customer complained", input.DriverNotes)
	assert.Len(t, input.EventTimeline, 1)
	assert.Equal(t, "packed", input.EventTimeline[0].FromStatus)
	assert.Equal(t, "shipped", input.EventTimeline[0].ToStatus)
}

func TestExceptionAnalyzer_StructuralRuleMatch(t *testing.T) {
	// Setup structural exception context: Duplicate Event
	now := time.Now()
	aiCtx := &models.AIContext{
		OrderID:       1001,
		CreatedAt:     now.Add(-5 * time.Hour),
		CurrentStatus: models.ORDER_STATUS_PAID,
		TotalAmount:   100000,
		Events: []models.AIEvent{
			{
				EventAt:        now.Add(-4 * time.Hour),
				PreviousStatus: models.ORDER_STATUS_CREATED,
				NewStatus:      models.ORDER_STATUS_PAID,
				UpdatedBy:      "admin_1",
			},
			{
				EventAt:        now.Add(-3 * time.Hour),
				PreviousStatus: models.ORDER_STATUS_CREATED,
				NewStatus:      models.ORDER_STATUS_PAID, // Duplicate event!
				UpdatedBy:      "admin_1",
				DriverNote:     func(s string) *string { return &s }("giao hàng thành công tận nơi"), // valid note but structural rules should skip AI
			},
		},
	}

	analyzer := ai.NewExceptionAnalyzer(nil, ai.ExceptionAnalyzerConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	result, err := analyzer.Analyze(context.Background(), aiCtx)
	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, constant.FallbackReasonStructuralRuleMatch, result.FallbackReason)
	assert.Equal(t, "DUPLICATE_EVENT", result.ExceptionType)
}

func TestExceptionAnalyzer_SpamNotes(t *testing.T) {
	tests := []struct {
		name string
		note string
	}{
		{"Too short", "ok"},
		{"Too short 2", "đã giao"},
		{"Low letter ratio", "123 456 7890"},
		{"Low letter ratio 2", "!!! @@@ ###"},
		{"Too few words", "xe hỏng"},
		{"Too few words 2", "hàng mất"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aiCtx := newTestAIContextWithDriverNote(tt.note)
			analyzer := ai.NewExceptionAnalyzer(nil, ai.ExceptionAnalyzerConfig{
				AIEnabled: true,
				AITimeout: 10 * time.Second,
			})

			result, err := analyzer.Analyze(context.Background(), aiCtx)
			require.NoError(t, err)
			assert.True(t, result.FallbackUsed)
			assert.Equal(t, constant.FallbackReasonNoteNotActionable, result.FallbackReason)
		})
	}
}

func TestExceptionAnalyzer_ActionableNote_CallsAI(t *testing.T) {
	// A valid, actionable note should pass the spam gate and call AI
	adapter := &mockAdapter{
		output: dto_ai.AIAnalysisResult{
			ExceptionType:      "DELIVERY_FAILURE",
			Severity:           "HIGH",
			LikelyReason:       "Vehicle broke down",
			InternalNextAction: "Reschedule",
			ConfidenceScore:    0.9,
		},
		outputStr: `{"exception_type":"DELIVERY_FAILURE"}`,
	}
	analyzer := ai.NewExceptionAnalyzer(adapter, ai.ExceptionAnalyzerConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	aiCtx := newTestAIContextWithDriverNote("giao hàng thất bại vì hỏng xe giữa đường")
	result, err := analyzer.Analyze(context.Background(), aiCtx)
	require.NoError(t, err)
	assert.False(t, result.FallbackUsed)
	assert.Equal(t, "DELIVERY_FAILURE", result.ExceptionType)
}
