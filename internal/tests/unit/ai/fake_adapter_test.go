package ai_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"main/internal/ai"
	dto_ai "main/internal/dto/ai"
	dto_api "main/internal/dto/api"
)

func TestFakeAIAdapter_AnalyzeException(t *testing.T) {
	expectedOutput := dto_ai.AIAnalysisResult{
		Severity:           "HIGH",
		LikelyReason:       "Traffic congestion in metropolitan area",
		InternalNextAction: "Assign backup driver",
		ConfidenceScore:    0.95,
	}

	expectedSummary := dto_api.ReportSummaryOutput{
		Summary:     "Operational efficiency is stable with some traffic delay exceptions.",
		Highlights:  []string{"High traffic delay"},
		Suggestions: []string{"Reroute delivery path"},
	}

	input := dto_ai.ExceptionPromptContext{
		OrderID:         12345,
		CurrentStatus:   "shipped",
		TotalAmount:     500000,
		CustomerName:    "John Doe",
		ShippingAddress: "123 Main St",
		CreatedAt:       "2023-10-27T10:00:00Z",
		DriverNotes:     "Driver got stuck in traffic",
		EventTimeline:   nil,
	}

	tests := []struct {
		name      string
		setup     func(f *ai.FakeAIAdapter)
		assertErr func(t *testing.T, err error)
		assertOut func(t *testing.T, output dto_ai.AIAnalysisResult, rawText string)
	}{
		{
			name: "Happy Path - Normal AI response",
			setup: func(f *ai.FakeAIAdapter) {
				// No special configuration
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertOut: func(t *testing.T, output dto_ai.AIAnalysisResult, rawText string) {
				assert.Equal(t, expectedOutput.Severity, output.Severity)
				assert.Contains(t, rawText, expectedOutput.Severity)
				t.Logf("[HAPPY PATH OUTPUT] Output: %+v", output)
			},
		},
		{
			name: "Failure Path - AI disabled",
			setup: func(f *ai.FakeAIAdapter) {
				f.SimulateAIDisabled = true
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Equal(t, ai.ErrAIDisabled, err)
				t.Logf("[AI DISABLED ERROR] Gặp lỗi giả lập khi AI disabled: %v", err)
			},
			assertOut: func(t *testing.T, output dto_ai.AIAnalysisResult, rawText string) {
				assert.Empty(t, output.ExceptionType)
				assert.Empty(t, rawText)
			},
		},
		{
			name: "Failure Path - Timeout",
			setup: func(f *ai.FakeAIAdapter) {
				f.SimulateTimeout = true
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
				t.Logf("[TIMEOUT ERROR] Gặp lỗi giả lập AI Timeout: %v", err)
			},
			assertOut: func(t *testing.T, output dto_ai.AIAnalysisResult, rawText string) {
				assert.Empty(t, output.ExceptionType)
				assert.Empty(t, rawText)
			},
		},
		{
			name: "Failure Path - Invalid response",
			setup: func(f *ai.FakeAIAdapter) {
				f.SimulateInvalidResponse = true
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err) // invalid response now returns an error
			},
			assertOut: func(t *testing.T, output dto_ai.AIAnalysisResult, rawText string) {
				assert.Empty(t, output.ExceptionType)
				assert.Equal(t, "{invalid-json}", rawText)
				t.Logf("[INVALID RESPONSE OUTPUT] Error returned for invalid response")
			},
		},
		{
			name: "Failure Path - Low confidence",
			setup: func(f *ai.FakeAIAdapter) {
				f.SimulateLowConfidence = true
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertOut: func(t *testing.T, output dto_ai.AIAnalysisResult, rawText string) {
				assert.Equal(t, 0.3, output.ConfidenceScore)
				assert.Contains(t, rawText, `"confidence_score":0.3`)
				t.Logf("[LOW CONFIDENCE OUTPUT] Output with low confidence: %+v", output)
			},
		},
		{
			name: "Failure Path - Connection error",
			setup: func(f *ai.FakeAIAdapter) {
				f.SimulateConnectionError = true
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "connection reset by peer")
				t.Logf("[CONNECTION ERROR] Gặp lỗi mất kết nối mạng: %v", err)
			},
			assertOut: func(t *testing.T, output dto_ai.AIAnalysisResult, rawText string) {
				assert.Empty(t, output.ExceptionType)
				assert.Empty(t, rawText)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := ai.NewFakeAIAdapter(expectedOutput, expectedSummary)
			tt.setup(adapter)

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			output, rawText, err := adapter.AnalyzeException(ctx, input)
			tt.assertErr(t, err)
			tt.assertOut(t, output, rawText)
		})
	}
}

func TestFakeAIAdapter_SummarizeReport(t *testing.T) {
	expectedOutput := dto_ai.AIAnalysisResult{
		Severity: "HIGH",
	}

	expectedSummary := dto_api.ReportSummaryOutput{
		Summary:     "Weekly summary details.",
		Highlights:  []string{"Delay on Monday"},
		Suggestions: []string{"Optimize route"},
	}

	tests := []struct {
		name      string
		setup     func(f *ai.FakeAIAdapter)
		assertErr func(t *testing.T, err error)
		assertOut func(t *testing.T, output dto_api.ReportSummaryOutput)
	}{
		{
			name: "Happy Path - Normal Summary",
			setup: func(f *ai.FakeAIAdapter) {
				// No special configuration
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertOut: func(t *testing.T, output dto_api.ReportSummaryOutput) {
				assert.Equal(t, expectedSummary.Summary, output.Summary)
				assert.Equal(t, expectedSummary.Highlights, output.Highlights)
				t.Logf("[HAPPY PATH SUMMARY OUTPUT] Summary: %s, Highlights: %v, Suggestions: %v",
					output.Summary, output.Highlights, output.Suggestions)
			},
		},
		{
			name: "Failure Path - AI disabled",
			setup: func(f *ai.FakeAIAdapter) {
				f.SimulateAIDisabled = true
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Equal(t, ai.ErrAIDisabled, err)
				t.Logf("[AI DISABLED ERROR ON SUMMARY] Lỗi khi AI bị disabled: %v", err)
			},
			assertOut: func(t *testing.T, output dto_api.ReportSummaryOutput) {
				assert.Empty(t, output)
			},
		},
		{
			name: "Failure Path - Timeout",
			setup: func(f *ai.FakeAIAdapter) {
				f.SimulateTimeout = true
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				t.Logf("[TIMEOUT ERROR ON SUMMARY] Lỗi AI Timeout khi tóm tắt: %v", err)
			},
			assertOut: func(t *testing.T, output dto_api.ReportSummaryOutput) {
				assert.Empty(t, output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := ai.NewFakeAIAdapter(expectedOutput, expectedSummary)
			tt.setup(adapter)

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			output, err := adapter.SummarizeReport(ctx, expectedOutput)
			tt.assertErr(t, err)
			tt.assertOut(t, output)
		})
	}
}

func TestFakeAIAdapter_Ping(t *testing.T) {
	expectedOutput := dto_ai.AIAnalysisResult{}
	expectedSummary := dto_api.ReportSummaryOutput{}

	tests := []struct {
		name      string
		setup     func(f *ai.FakeAIAdapter)
		assertErr func(t *testing.T, err error)
	}{
		{
			name: "Happy Path - Success Ping",
			setup: func(f *ai.FakeAIAdapter) {
				// No special configuration
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
				t.Logf("[PING SUCCESS] Ping thành công tới AI Provider.")
			},
		},
		{
			name: "Failure Path - AI disabled",
			setup: func(f *ai.FakeAIAdapter) {
				f.SimulateAIDisabled = true
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Equal(t, ai.ErrAIDisabled, err)
				t.Logf("[PING ERROR - AI DISABLED] Ping thất bại vì AI bị disable: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := ai.NewFakeAIAdapter(expectedOutput, expectedSummary)
			tt.setup(adapter)

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			err := adapter.Ping(ctx)
			tt.assertErr(t, err)
		})
	}
}
