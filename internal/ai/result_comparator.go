package ai

import (
	"main/internal/dto"
)

// CaseResult holds the evaluation result of a single test case.
type CaseResult struct {
	CaseID           string
	IsPassed         bool
	ExpectedType     string
	ActualType       string
	ExpectedSeverity string
	ActualSeverity   string
	FallbackUsed     bool
	LatencyMs        int
	ErrorMessage     string
}

// RunMetrics holds the aggregated metrics for an entire evaluation batch.
type RunMetrics struct {
	TotalCases    int
	PassedCases   int
	FailedCases   int
	FallbackCount int
	AccuracyRate  float64
	AvgLatencyMs  int
}

// CompareResult compares the ground truth (expected) with the actual AI analysis result.
// It requires an exact match on both ExceptionType and Severity to be considered PASSED.
func CompareResult(expected dto.EvaluationCase, actual *AnalysisResult) *CaseResult {
	if actual == nil {
		return &CaseResult{
			CaseID:           expected.CaseID,
			IsPassed:         false,
			ExpectedType:     expected.ExpectedExceptionType,
			ExpectedSeverity: expected.ExpectedSeverity,
			ErrorMessage:     "actual result is nil",
		}
	}

	isPassed := expected.ExpectedExceptionType == actual.ExceptionType &&
		expected.ExpectedSeverity == actual.Severity

	return &CaseResult{
		CaseID:           expected.CaseID,
		IsPassed:         isPassed,
		ExpectedType:     expected.ExpectedExceptionType,
		ActualType:       actual.ExceptionType,
		ExpectedSeverity: expected.ExpectedSeverity,
		ActualSeverity:   actual.Severity,
		FallbackUsed:     actual.FallbackUsed,
		LatencyMs:        actual.DurationMs,
		ErrorMessage:     "",
	}
}

// CalculateRunMetrics aggregates a slice of CaseResults into overall RunMetrics.
func CalculateRunMetrics(results []*CaseResult) RunMetrics {
	metrics := RunMetrics{
		TotalCases: len(results),
	}

	if metrics.TotalCases == 0 {
		return metrics
	}

	totalLatency := 0

	for _, res := range results {
		if res.IsPassed {
			metrics.PassedCases++
		} else {
			metrics.FailedCases++
		}

		if res.FallbackUsed {
			metrics.FallbackCount++
		}

		totalLatency += res.LatencyMs
	}

	metrics.AccuracyRate = float64(metrics.PassedCases) / float64(metrics.TotalCases) * 100.0
	metrics.AvgLatencyMs = totalLatency / metrics.TotalCases

	return metrics
}
