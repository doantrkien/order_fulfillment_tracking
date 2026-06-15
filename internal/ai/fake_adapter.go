package ai

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"main/internal/dto"
)

// ErrAIDisabled is returned when the AI module is simulated as disabled.
var ErrAIDisabled = errors.New("AI module is disabled")

// FakeAIAdapter implements AIAdapter for testing purposes.
type FakeAIAdapter struct {
	SimulateAIDisabled      bool
	SimulateTimeout         bool
	SimulateInvalidResponse bool
	SimulateLowConfidence   bool
	SimulateConnectionError bool
	ExpectedOutput          dto.ExceptionOutput
	ExpectedSummary         dto.ReportSummaryOutput
}

// NewFakeAIAdapter creates a new instance of FakeAIAdapter.
func NewFakeAIAdapter(
	expected dto.ExceptionOutput,
	summary dto.ReportSummaryOutput,
) *FakeAIAdapter {
	return &FakeAIAdapter{
		ExpectedOutput:  expected,
		ExpectedSummary: summary,
	}
}

// AnalyzeException simulates AI analysis for order exception detection.
func (f *FakeAIAdapter) AnalyzeException(
	ctx context.Context,
	input dto.ExceptionInput,
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
		// Output structured JSON representing a low confidence score.
		lowConfJSON := `{
			"exception_type": "STUCK_ORDER",
			"severity": "HIGH",
			"likely_reason": "Order stuck in status packed too long",
			"internal_next_action": "Contact warehouse manager",
			"suggestion": "Contact warehouse manager",
			"should_alert": true,
			"confidence_score": 0.3
		}`
		return lowConfJSON, nil
	}

	// Default: Return the ExpectedOutput marshaled to string (JSON) to pass parsing and validation.
	type TempOutput struct {
		ExceptionType      string  `json:"exception_type"`
		Severity           string  `json:"severity"`
		LikelyReason       string  `json:"likely_reason"`
		InternalNextAction string  `json:"internal_next_action"`
		Suggestion         string  `json:"suggestion"`
		ShouldAlert        bool    `json:"should_alert"`
		ConfidenceScore    float64 `json:"confidence_score"`
	}
	tOut := TempOutput{
		ExceptionType:      f.ExpectedOutput.ExceptionType,
		Severity:           f.ExpectedOutput.Severity,
		LikelyReason:       f.ExpectedOutput.LikelyReason,
		InternalNextAction: f.ExpectedOutput.InternalNextAction,
		Suggestion:         f.ExpectedOutput.Suggestion,
		ShouldAlert:        f.ExpectedOutput.ShouldAlert,
		ConfidenceScore:    f.ExpectedOutput.Confidence,
	}
	if tOut.ConfidenceScore == 0 {
		tOut.ConfidenceScore = 0.95 // Default high confidence if not set
	}
	if tOut.ExceptionType == "" {
		tOut.ExceptionType = "STUCK_ORDER"
	}
	if tOut.Severity == "" {
		tOut.Severity = "HIGH"
	}
	if tOut.LikelyReason == "" {
		tOut.LikelyReason = "Default reason"
	}
	if tOut.InternalNextAction == "" {
		tOut.InternalNextAction = "Default next action"
	}

	data, err := json.Marshal(tOut)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DraftCustomerUpdate simulates AI drafting a customer update message.
func (f *FakeAIAdapter) DraftCustomerUpdate(
	ctx context.Context,
	input dto.CustomerUpdateDraftInput,
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
	input dto.ExceptionOutput,
) (dto.ReportSummaryOutput, error) {
	if f.SimulateAIDisabled {
		return dto.ReportSummaryOutput{}, ErrAIDisabled
	}

	if f.SimulateTimeout {
		select {
		case <-ctx.Done():
			return dto.ReportSummaryOutput{}, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return dto.ReportSummaryOutput{}, context.DeadlineExceeded
		}
	}

	return f.ExpectedSummary, nil
}

// Ping simulates checking connectivity/credentials for the AI adapter.
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
