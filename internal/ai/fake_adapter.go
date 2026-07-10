package ai

import (
	"context"
	"encoding/json"
	"errors"
	dto_ai "main/internal/dto/ai"
	dto_api "main/internal/dto/api"
	"time"
)

var ErrAIDisabled = errors.New("AI module is disabled")

// FakeAIAdapter implements AIAdapter for testing purposes.
type FakeAIAdapter struct {
	SimulateAIDisabled      bool
	SimulateTimeout         bool
	SimulateInvalidResponse bool
	SimulateLowConfidence   bool
	SimulateConnectionError bool
	ExpectedOutput          dto_ai.AIAnalysisResult
	ExpectedSummary         dto_api.ReportSummaryOutput
}

// NewFakeAIAdapter creates a new instance of FakeAIAdapter.
func NewFakeAIAdapter(
	expected dto_ai.AIAnalysisResult,
	summary dto_api.ReportSummaryOutput,
) *FakeAIAdapter {
	return &FakeAIAdapter{
		ExpectedOutput:  expected,
		ExpectedSummary: summary,
	}
}

// AnalyzeException simulates AI analysis for order exception detection.
func (f *FakeAIAdapter) AnalyzeException(
	ctx context.Context,
	input dto_ai.ExceptionPromptContext,
) (dto_ai.AIAnalysisResult, string, error) {
	if f.SimulateAIDisabled {
		return dto_ai.AIAnalysisResult{}, "", ErrAIDisabled
	}

	if f.SimulateConnectionError {
		return dto_ai.AIAnalysisResult{}, "", errors.New("connection reset by peer")
	}

	if f.SimulateTimeout {
		select {
		case <-ctx.Done():
			return dto_ai.AIAnalysisResult{}, "", ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return dto_ai.AIAnalysisResult{}, "", context.DeadlineExceeded
		}
	}

	if f.SimulateInvalidResponse {
		return dto_ai.AIAnalysisResult{}, "{invalid-json}", errors.New("AI response is not valid JSON")
	}

	if f.SimulateLowConfidence {
		out := dto_ai.AIAnalysisResult{
			ExceptionType:      "STUCK_ORDER",
			Severity:           "HIGH",
			LikelyReason:       "Order stuck in status packed too long",
			InternalNextAction: "Contact warehouse manager",
			ConfidenceScore:    0.3,
		}
		rawTextBytes, _ := json.Marshal(out)
		return out, string(rawTextBytes), nil
	}

	// Default: Return the ExpectedOutput with sensible defaults.
	out := f.ExpectedOutput
	if out.ConfidenceScore == 0 {
		out.ConfidenceScore = 0.95 // Default high confidence if not set
	}
	if out.ExceptionType == "" {
		out.ExceptionType = "STUCK_ORDER"
	}
	if out.Severity == "" {
		out.Severity = "HIGH"
	}
	if out.LikelyReason == "" {
		out.LikelyReason = "Default reason"
	}
	if out.InternalNextAction == "" {
		out.InternalNextAction = "Default next action"
	}

	rawTextBytes, _ := json.Marshal(out)
	return out, string(rawTextBytes), nil
}

// DraftCustomerUpdate simulates AI drafting a customer update message.
func (f *FakeAIAdapter) DraftCustomerUpdate(
	ctx context.Context,
	input dto_ai.CustomerUpdateDraftInput,
) (string, error) {
	if f.SimulateAIDisabled {
		return "", ErrAIDisabled
	}

	if f.SimulateConnectionError {
		return "", errors.New("connection reset by peer")
	}

	if f.SimulateTimeout {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return "", context.DeadlineExceeded
		}
	}

	if f.SimulateInvalidResponse {
		return "{invalid-json}", nil
	}

	if f.SimulateLowConfidence {
		lowConfJSON := `{
			"customer_update_draft": "We are looking into your order.",
			"confidence_score": 0.3
		}`
		return lowConfJSON, nil
	}

	// Default success
	successJSON := `{
		"customer_update_draft": "Your order is currently delayed due to weather conditions. We apologize for the inconvenience and will update you soon.",
		"confidence_score": 0.95
	}`
	return successJSON, nil
}

// SummarizeReport simulates AI report summarization.
func (f *FakeAIAdapter) SummarizeReport(
	ctx context.Context,
	input dto_ai.AIAnalysisResult,
) (dto_api.ReportSummaryOutput, error) {
	if f.SimulateAIDisabled {
		return dto_api.ReportSummaryOutput{}, ErrAIDisabled
	}

	if f.SimulateTimeout {
		select {
		case <-ctx.Done():
			return dto_api.ReportSummaryOutput{}, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return dto_api.ReportSummaryOutput{}, context.DeadlineExceeded
		}
	}

	return f.ExpectedSummary, nil
}

func (f *FakeAIAdapter) Ping(ctx context.Context) error {
	if f.SimulateAIDisabled {
		return ErrAIDisabled
	}

	if f.SimulateTimeout {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return context.DeadlineExceeded
		}
	}

	return nil
}

func (f *FakeAIAdapter) ReloadKnowledge(_ context.Context) error {
	return nil
}
