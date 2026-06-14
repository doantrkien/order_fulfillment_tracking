package ai_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"main/internal/ai"
	"main/internal/dto"
)

func TestFakeAIAdapter_AnalyzeException(t *testing.T) {
	expectedOutput := dto.ExceptionOutput{
		Severity:     "HIGH",
		LikelyReason: "Traffic congestion in metropolitan area",
		Suggestion:   "Assign backup driver",
		ShouldAlert:  true,
		Confidence:   0.95,
	}

	expectedSummary := dto.ReportSummaryOutput{
		Summary:     "Operational efficiency is stable with some traffic delay exceptions.",
		Highlights:  []string{"High traffic delay"},
		Suggestions: []string{"Reroute delivery path"},
	}

	input := dto.ExceptionInput{
		OrderID:         12345,
		CurrentStatus:   "shipped",
		TotalAmount:     500000,
		CustomerName:    "John Doe",
		ShippingAddress: "123 Main St",
		CreatedAt:       "2023-10-27T10:00:00Z",
		ErrorMessage:    "Driver got stuck in traffic",
		EventHistory:    nil,
	}

	tests := []struct {
		name      string
		setup     func(f *ai.FakeAIAdapter)
		assertErr func(t *testing.T, err error)
		assertOut func(t *testing.T, output string)
	}{
		{
			name: "Happy Path - Normal AI response",
			setup: func(f *ai.FakeAIAdapter) {
				// No special configuration
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertOut: func(t *testing.T, output string) {
				assert.Contains(t, output, expectedOutput.Severity)
				t.Logf("[HAPPY PATH OUTPUT] Raw JSON Output: %s", output)
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
			assertOut: func(t *testing.T, output string) {
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
				assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
				t.Logf("[TIMEOUT ERROR] Gặp lỗi giả lập AI Timeout: %v", err)
			},
			assertOut: func(t *testing.T, output string) {
				assert.Empty(t, output)
			},
		},
		{
			name: "Failure Path - Invalid response",
			setup: func(f *ai.FakeAIAdapter) {
				f.SimulateInvalidResponse = true
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err) // in the new FakeAIAdapter, invalid response does not return an error immediately, it returns invalid JSON string
			},
			assertOut: func(t *testing.T, output string) {
				assert.Equal(t, "{invalid-json}", output)
				t.Logf("[INVALID RESPONSE OUTPUT] Gặp chuỗi JSON không hợp lệ: %s", output)
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
			assertOut: func(t *testing.T, output string) {
				assert.Contains(t, output, `"confidence_score": 0.3`)
				t.Logf("[LOW CONFIDENCE OUTPUT] JSON với điểm tin cậy thấp: %s", output)
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
			assertOut: func(t *testing.T, output string) {
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

			output, err := adapter.AnalyzeException(ctx, input)
			tt.assertErr(t, err)
			tt.assertOut(t, output)
		})
	}
}

func TestFakeAIAdapter_SummarizeReport(t *testing.T) {
	expectedOutput := dto.ExceptionOutput{
		Severity: "HIGH",
	}

	expectedSummary := dto.ReportSummaryOutput{
		Summary:     "Weekly summary details.",
		Highlights:  []string{"Delay on Monday"},
		Suggestions: []string{"Optimize route"},
	}

	tests := []struct {
		name      string
		setup     func(f *ai.FakeAIAdapter)
		assertErr func(t *testing.T, err error)
		assertOut func(t *testing.T, output dto.ReportSummaryOutput)
	}{
		{
			name: "Happy Path - Normal Summary",
			setup: func(f *ai.FakeAIAdapter) {
				// No special configuration
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertOut: func(t *testing.T, output dto.ReportSummaryOutput) {
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
			assertOut: func(t *testing.T, output dto.ReportSummaryOutput) {
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
			assertOut: func(t *testing.T, output dto.ReportSummaryOutput) {
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
	expectedOutput := dto.ExceptionOutput{}
	expectedSummary := dto.ReportSummaryOutput{}

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
