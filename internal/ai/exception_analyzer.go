package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/dto"
	"main/internal/models"
	"time"
)

// Fallback reason constants used in audit logging.
const (
	FallbackReasonDisabled        = "ai_disabled"
	FallbackReasonTimeout         = "ai_timeout"
	FallbackReasonConnectionError = "ai_connection_error"
	FallbackReasonInvalidResponse = "ai_invalid_response"
	FallbackReasonLowConfidence   = "ai_confidence_below_threshold"
)

// AnalysisResult is the unified output produced by the ExceptionAnalyzer.
// Both AI and rule-based paths produce this same type.
type AnalysisResult struct {
	ExceptionType      string
	Severity           string
	LikelyReason       string
	InternalNextAction string
	ConfidenceScore    float64
	FallbackUsed       bool
	FallbackReason     string
	DurationMs         int
	RawResponse        string // Raw AI response (empty when fallback)
}

// ExceptionAnalyzerConfig holds runtime configuration for the analyzer.
type ExceptionAnalyzerConfig struct {
	AIEnabled bool
	AITimeout time.Duration
}

// ExceptionAnalyzer orchestrates AI exception analysis with automatic
// fallback to deterministic rule-based detection.
type ExceptionAnalyzer struct {
	adapter AIAdapter
	config  ExceptionAnalyzerConfig
}

// NewExceptionAnalyzer creates an analyzer with the given AI adapter and config.
// adapter may be nil when AI is disabled.
func NewExceptionAnalyzer(adapter AIAdapter, config ExceptionAnalyzerConfig) *ExceptionAnalyzer {
	return &ExceptionAnalyzer{
		adapter: adapter,
		config:  config,
	}
}

// Analyze is the main entry point. It follows this flow:
//  1. Check if AI is enabled → if not, fallback immediately
//  2. Call AI adapter with timeout context
//  3. Check confidence threshold on the result
//  4. Falls back to rule-based if any step fails
//
// This method never returns an error for AI failures — it always falls back.
// Errors are only returned for genuine infrastructure issues (e.g., invalid input).
func (ea *ExceptionAnalyzer) Analyze(ctx context.Context, aiCtx *models.AIContext, notes string) (*AnalysisResult, error) {
	now := time.Now()

	// Step 1: Check if AI is disabled
	if !ea.config.AIEnabled {
		return ea.fallback(aiCtx, FallbackReasonDisabled, 0, now), nil
	}

	// Step 2: Build input and call AI adapter with timeout
	input := buildExceptionInput(aiCtx, notes)

	start := time.Now()
	aiCtxTimeout, cancel := context.WithTimeout(ctx, ea.config.AITimeout)
	defer cancel()

	aiOutput, err := ea.adapter.AnalyzeException(aiCtxTimeout, input)
	durationMs := int(time.Since(start).Milliseconds())

	// Step 3: Handle AI call errors (timeout, connection, etc.)
	if err != nil {
		reason := ClassifyError(err)
		return ea.fallback(aiCtx, reason, durationMs, now), nil
	}

	// Step 4: Validate response by marshaling back to JSON and running schema validation
	rawJSON, marshalErr := json.Marshal(aiOutput)
	if marshalErr != nil {
		return ea.fallbackWithRaw(aiCtx, FallbackReasonInvalidResponse, durationMs, "", now), nil
	}

	rawStr := string(rawJSON)
	validated, validationErr := ParseAndValidateAIOutput(rawStr)
	if validationErr != nil {
		return ea.fallbackWithRaw(aiCtx, FallbackReasonInvalidResponse, durationMs, rawStr, now), nil
	}

	// Step 5: Check confidence threshold
	if shouldFallback, _ := ShouldFallback(validated); shouldFallback {
		return ea.fallbackWithRaw(aiCtx, FallbackReasonLowConfidence, durationMs, rawStr, now), nil
	}

	// Step 6: AI succeeded — return the validated result
	return &AnalysisResult{
		ExceptionType:      validated.ExceptionType,
		Severity:           validated.Severity,
		LikelyReason:       validated.LikelyReason,
		InternalNextAction: validated.InternalNextAction,
		ConfidenceScore:    validated.ConfidenceScore,
		FallbackUsed:       false,
		DurationMs:         durationMs,
		RawResponse:        rawStr,
	}, nil
}

// fallback runs the rule-based analyzer and wraps the result.
func (ea *ExceptionAnalyzer) fallback(aiCtx *models.AIContext, reason string, durationMs int, now time.Time) *AnalysisResult {
	return ea.fallbackWithRaw(aiCtx, reason, durationMs, "", now)
}

// fallbackWithRaw runs the rule-based analyzer, preserving the raw AI response for audit.
func (ea *ExceptionAnalyzer) fallbackWithRaw(aiCtx *models.AIContext, reason string, durationMs int, rawResponse string, now time.Time) *AnalysisResult {
	ruleResult := AnalyzeByRules(aiCtx, now)

	if ruleResult == nil {
		// No exception detected by rules either
		return &AnalysisResult{
			ExceptionType:      "OTHER",
			Severity:           "LOW",
			LikelyReason:       "No specific exception detected by rule-based analysis",
			InternalNextAction: "No action required. Monitor the order for further changes.",
			ConfidenceScore:    1.0,
			FallbackUsed:       true,
			FallbackReason:     reason,
			DurationMs:         durationMs,
			RawResponse:        rawResponse,
		}
	}

	return &AnalysisResult{
		ExceptionType:      ruleResult.ExceptionType,
		Severity:           ruleResult.Severity,
		LikelyReason:       ruleResult.LikelyReason,
		InternalNextAction: ruleResult.InternalNextAction,
		ConfidenceScore:    ruleResult.ConfidenceScore,
		FallbackUsed:       true,
		FallbackReason:     reason,
		DurationMs:         durationMs,
		RawResponse:        rawResponse,
	}
}

// buildExceptionInput converts the internal AIContext to the adapter's input DTO.
func buildExceptionInput(aiCtx *models.AIContext, notes string) dto.ExceptionInput {
	eventHistory := make([]dto.EventRecord, 0, len(aiCtx.Events))
	for _, e := range aiCtx.Events {
		eventHistory = append(eventHistory, dto.EventRecord{
			FromStatus: string(e.PreviousStatus),
			ToStatus:   string(e.NewStatus),
			EventAt:    e.EventAt.Format(time.RFC3339),
			UpdatedBy:  e.UpdatedBy,
		})
	}

	return dto.ExceptionInput{
		OrderID:       aiCtx.OrderID,
		CurrentStatus: string(aiCtx.CurrentStatus),
		ErrorMessage:  notes,
		EventHistory:  eventHistory,
	}
}

// FormatFallbackReason returns a human-readable description of the fallback reason.
func FormatFallbackReason(reason string) string {
	switch reason {
	case FallbackReasonDisabled:
		return "AI feature is disabled via configuration"
	case FallbackReasonTimeout:
		return "AI service call timed out"
	case FallbackReasonConnectionError:
		return "Could not connect to AI service"
	case FallbackReasonInvalidResponse:
		return "AI returned an invalid or malformed response"
	case FallbackReasonLowConfidence:
		return "AI response confidence score was below threshold"
	default:
		return fmt.Sprintf("Unknown fallback reason: %s", reason)
	}
}
