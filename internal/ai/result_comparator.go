package ai

import (
	"main/internal/dto"
)

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

type RunMetrics struct {
	TotalCases    int
	PassedCases   int
	FailedCases   int
	FallbackCount int
	AccuracyRate  float64
	AvgLatencyMs  int
}

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
