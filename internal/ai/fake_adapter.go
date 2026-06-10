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
) (dto.ExceptionOutput, error) {
	if f.SimulateAIDisabled {
		return dto.ExceptionOutput{}, ErrAIDisabled
	}

	if f.SimulateTimeout {
		select {
		case <-ctx.Done():
			return dto.ExceptionOutput{}, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return dto.ExceptionOutput{}, context.DeadlineExceeded
		}
	}

	if f.SimulateInvalidResponse {
		var out dto.ExceptionOutput
		// Triggers an unmarshal error to simulate bad/garbage JSON.
		err := json.Unmarshal([]byte("{invalid-json}"), &out)
		return dto.ExceptionOutput{}, err
	}

	if f.SimulateLowConfidence {
		out := f.ExpectedOutput
		out.Confidence = 0.3
		return out, nil
	}

	return f.ExpectedOutput, nil
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
