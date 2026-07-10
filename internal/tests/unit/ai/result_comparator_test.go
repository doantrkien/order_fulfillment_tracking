package ai_test

import (
	"testing"

	"main/internal/ai"
	dto_ai "main/internal/dto/ai"

	"github.com/stretchr/testify/assert"
)

func TestCompareResult(t *testing.T) {
	tests := []struct {
		name     string
		expected dto_ai.EvaluationCase
		actual   *dto_ai.AIAnalysisResult
		validate func(t *testing.T, res *ai.CaseResult)
	}{
		{
			name: "Exact match -> PASSED",
			expected: dto_ai.EvaluationCase{
				CaseID:                "case_01",
				ExpectedExceptionType: "STUCK_ORDER",
				ExpectedSeverity:      "HIGH",
			},
			actual: &dto_ai.AIAnalysisResult{
				ExceptionType: "STUCK_ORDER",
				Severity:      "HIGH",
				FallbackUsed:  false,
				DurationMs:    1500,
			},
			validate: func(t *testing.T, res *ai.CaseResult) {
				assert.True(t, res.IsPassed)
				assert.Equal(t, "STUCK_ORDER", res.ActualType)
				assert.Equal(t, "HIGH", res.ActualSeverity)
				assert.False(t, res.FallbackUsed)
				assert.Equal(t, 1500, res.LatencyMs)
			},
		},
		{
			name: "Type match but Severity mismatch -> FAILED",
			expected: dto_ai.EvaluationCase{
				CaseID:                "case_02",
				ExpectedExceptionType: "STUCK_ORDER",
				ExpectedSeverity:      "CRITICAL",
			},
			actual: &dto_ai.AIAnalysisResult{
				ExceptionType: "STUCK_ORDER",
				Severity:      "HIGH",
			},
			validate: func(t *testing.T, res *ai.CaseResult) {
				assert.False(t, res.IsPassed)
			},
		},
		{
			name: "Type mismatch -> FAILED",
			expected: dto_ai.EvaluationCase{
				CaseID:                "case_03",
				ExpectedExceptionType: "SKIPPED_STATUS",
				ExpectedSeverity:      "HIGH",
			},
			actual: &dto_ai.AIAnalysisResult{
				ExceptionType: "STUCK_ORDER",
				Severity:      "HIGH",
			},
			validate: func(t *testing.T, res *ai.CaseResult) {
				assert.False(t, res.IsPassed)
			},
		},
		{
			name: "Fallback used but matches expected -> PASSED",
			expected: dto_ai.EvaluationCase{
				CaseID:                "case_04",
				ExpectedExceptionType: "OTHER",
				ExpectedSeverity:      "LOW",
			},
			actual: &dto_ai.AIAnalysisResult{
				ExceptionType: "OTHER",
				Severity:      "LOW",
				FallbackUsed:  true,
			},
			validate: func(t *testing.T, res *ai.CaseResult) {
				assert.True(t, res.IsPassed)
				assert.True(t, res.FallbackUsed)
			},
		},
		{
			name: "Nil actual result -> FAILED",
			expected: dto_ai.EvaluationCase{
				CaseID:                "case_05",
				ExpectedExceptionType: "OTHER",
				ExpectedSeverity:      "LOW",
			},
			actual: nil,
			validate: func(t *testing.T, res *ai.CaseResult) {
				assert.False(t, res.IsPassed)
				assert.NotEmpty(t, res.ErrorMessage)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ai.CompareResult(tt.expected, tt.actual)
			assert.Equal(t, tt.expected.CaseID, res.CaseID)
			tt.validate(t, res)
		})
	}
}

func TestCalculateRunMetrics(t *testing.T) {
	results := []*ai.CaseResult{
		{IsPassed: true, FallbackUsed: false, LatencyMs: 1000},
		{IsPassed: true, FallbackUsed: true, LatencyMs: 500},
		{IsPassed: false, FallbackUsed: false, LatencyMs: 1500},
		{IsPassed: false, FallbackUsed: true, LatencyMs: 200},
		{IsPassed: true, FallbackUsed: false, LatencyMs: 800},
	}

	metrics := ai.CalculateRunMetrics(results)

	assert.Equal(t, 5, metrics.TotalCases)
	assert.Equal(t, 3, metrics.PassedCases)
	assert.Equal(t, 2, metrics.FailedCases)
	assert.Equal(t, 2, metrics.FallbackCount)
	assert.Equal(t, 60.0, metrics.AccuracyRate) // 3/5 = 60%
	assert.Equal(t, 800, metrics.AvgLatencyMs)  // (1000+500+1500+200+800)/5 = 4000/5 = 800
}

func TestCalculateRunMetrics_Empty(t *testing.T) {
	metrics := ai.CalculateRunMetrics([]*ai.CaseResult{})
	assert.Equal(t, 0, metrics.TotalCases)
	assert.Equal(t, 0, metrics.PassedCases)
	assert.Equal(t, 0.0, metrics.AccuracyRate)
}
