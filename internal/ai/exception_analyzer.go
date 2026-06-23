package ai

import (
	"context"
	"fmt"
	"main/internal/dto"
	"main/internal/models"
	"strings"
	"time"
)

// Fallback reason constants used in audit logging.
const (
	FallbackReasonDisabled        = "ai_disabled"
	FallbackReasonNoDriverNote    = "no_driver_note" // rule-based result, AI not needed
	FallbackReasonTimeout         = "ai_timeout"
	FallbackReasonConnectionError = "ai_connection_error"
	FallbackReasonInvalidResponse = "ai_invalid_response"
	FallbackReasonLowConfidence   = "ai_confidence_below_threshold"
)

type AnalysisResult struct {
	ExceptionType      string
	Severity           string
	LikelyReason       string
	InternalNextAction string
	ConfidenceScore    float64
	FallbackUsed       bool
	FallbackReason     string
	DurationMs         int
	RawResponse        string
}

type ExceptionAnalyzerConfig struct {
	AIEnabled bool
	AITimeout time.Duration
}

type ExceptionAnalyzer struct {
	adapter AIAdapter
	config  ExceptionAnalyzerConfig
}

func NewExceptionAnalyzer(adapter AIAdapter, config ExceptionAnalyzerConfig) *ExceptionAnalyzer {
	return &ExceptionAnalyzer{
		adapter: adapter,
		config:  config,
	}
}

// Analyze runs the exception analysis pipeline:
//  1. Always run the rule-based engine first (fast, deterministic, no AI cost).
//  2. Check whether aiCtx.DriverNote (order-level) or any event carries a driver note.
//  3. Only call the AI when a driver note is present — it provides the rich
//     free-text context that makes AI analysis valuable.
//  4. If AI is disabled or no driver note exists, return the rule-based result.
func (ea *ExceptionAnalyzer) Analyze(ctx context.Context, aiCtx *models.AIContext, notes string) (*AnalysisResult, error) {
	now := time.Now()

	// ── Step 1: Rule-based engine (always runs) ──────────────────────────────
	ruleResult := AnalyzeByRules(aiCtx, now)

	// ── Step 2: Check for driver notes ───────────────────────────────────────
	hasDriverNote := hasAnyDriverNote(aiCtx)

	// ── Step 3: Skip AI when disabled or no driver note ──────────────────────
	if !ea.config.AIEnabled || !hasDriverNote {
		reason := FallbackReasonDisabled
		if ea.config.AIEnabled && !hasDriverNote {
			reason = FallbackReasonNoDriverNote
		}
		return ruleResultToAnalysis(ruleResult, reason, 0, ""), nil
	}

	// ── Step 4: Call AI (only when enabled AND driver note present) ───────────
	input := buildExceptionInput(aiCtx, notes)

	start := time.Now()
	aiCtxTimeout, cancel := context.WithTimeout(ctx, ea.config.AITimeout)
	defer cancel()

	aiOutput, rawText, err := ea.adapter.AnalyzeException(aiCtxTimeout, input)
	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		reason := ClassifyError(err)
		return ea.fallbackWithRaw(aiCtx, reason, durationMs, rawText, now), nil
	}

	if validationErr := ParseAndValidateAIOutput(&aiOutput); validationErr != nil {
		return ea.fallbackWithRaw(aiCtx, FallbackReasonInvalidResponse, durationMs, rawText, now), nil
	}

	if shouldFallback, _ := ShouldFallback(&aiOutput); shouldFallback {
		return ea.fallbackWithRaw(aiCtx, FallbackReasonLowConfidence, durationMs, rawText, now), nil
	}

	return &AnalysisResult{
		ExceptionType:      aiOutput.ExceptionType,
		Severity:           aiOutput.Severity,
		LikelyReason:       aiOutput.LikelyReason,
		InternalNextAction: aiOutput.InternalNextAction,
		ConfidenceScore:    aiOutput.ConfidenceScore,
		FallbackUsed:       false,
		DurationMs:         durationMs,
		RawResponse:        rawText,
	}, nil
}

// hasAnyDriverNote returns true if at least one event in the history carries
// a non-empty driver note.
func hasAnyDriverNote(aiCtx *models.AIContext) bool {
	for _, e := range aiCtx.Events {
		if e.DriverNote != nil && strings.TrimSpace(*e.DriverNote) != "" {
			return true
		}
	}
	return false
}

// ruleResultToAnalysis converts a RuleBasedResult (possibly nil) into an
// AnalysisResult. When ruleResult is nil no exception was detected by rules.
func ruleResultToAnalysis(ruleResult *RuleBasedResult, reason string, durationMs int, rawResponse string) *AnalysisResult {
	if ruleResult == nil {
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

// fallback runs the rule-based analyzer and wraps the result.
func (ea *ExceptionAnalyzer) fallback(aiCtx *models.AIContext, reason string, durationMs int, now time.Time) *AnalysisResult {
	return ea.fallbackWithRaw(aiCtx, reason, durationMs, "", now)
}

// fallbackWithRaw runs the rule-based analyzer, preserving the raw AI response for audit.
func (ea *ExceptionAnalyzer) fallbackWithRaw(aiCtx *models.AIContext, reason string, durationMs int, rawResponse string, now time.Time) *AnalysisResult {
	ruleResult := AnalyzeByRules(aiCtx, now)

	if ruleResult == nil {
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

	// Merge driver notes from all events in the DB with the API request note.
	// This gives AI the full picture of what drivers reported across the order lifecycle.
	allNotes := []string{}
	if strings.TrimSpace(notes) != "" {
		allNotes = append(allNotes, strings.TrimSpace(notes))
	}
	for _, e := range aiCtx.Events {
		if e.DriverNote != nil && strings.TrimSpace(*e.DriverNote) != "" {
			allNotes = append(allNotes, strings.TrimSpace(*e.DriverNote))
		}
	}
	errorMessage := strings.Join(allNotes, "; ")

	return dto.ExceptionInput{
		OrderID:         aiCtx.OrderID,
		CurrentStatus:   string(aiCtx.CurrentStatus),
		TotalAmount:     aiCtx.TotalAmount,
		CustomerName:    aiCtx.CustomerName,
		ShippingAddress: aiCtx.ShippingAddress,
		CreatedAt:       aiCtx.CreatedAt.Format(time.RFC3339),
		ErrorMessage:    errorMessage,
		EventHistory:    eventHistory,
	}
}

// FormatFallbackReason returns a human-readable description of the fallback reason.
func FormatFallbackReason(reason string) string {
	switch reason {
	case FallbackReasonDisabled:
		return "AI feature is disabled via configuration"
	case FallbackReasonNoDriverNote:
		return "No driver note present; rule-based result returned without calling AI"
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
