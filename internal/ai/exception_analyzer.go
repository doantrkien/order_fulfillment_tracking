package ai

import (
	"context"
	"fmt"
	"main/internal/dto"
	"main/internal/models"
	"time"
)

// Fallback reason constants used in audit logging.
const (
	FallbackReasonDisabled        = "ai_disabled"
	FallbackReasonTimeout         = "ai_timeout"
	FallbackReasonConnectionError  = "ai_connection_error"
	FallbackReasonInvalidResponse  = "ai_invalid_response"
	FallbackReasonLowConfidence    = "ai_confidence_below_threshold"
	FallbackReasonNoDriverNote     = "no_driver_note"
	FallbackReasonEarlyNoteIgnored = "early_note_ignored_to_save_cost"
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

func (ea *ExceptionAnalyzer) Analyze(ctx context.Context, aiCtx *models.AIContext, notes string) (*AnalysisResult, error) {
	now := time.Now()

	// Gate 1: AI disabled
	if !ea.config.AIEnabled {
		return ea.fallback(aiCtx, FallbackReasonDisabled, 0, now), nil
	}

	noteClassification := ClassifyDriverNote(notes)

	// Gate 2: No driver note
	if noteClassification.Category == DriverNoteCategoryNone {
		return ea.fallback(aiCtx, FallbackReasonNoDriverNote, 0, now), nil
	}

	slaBreached := IsSLABreached(aiCtx, now)

	// Gate 3: Early Note Ignored (Not breached SLA + General/Delay note)
	if !slaBreached && (noteClassification.Category == DriverNoteCategoryGeneral || noteClassification.Category == DriverNoteCategoryDelay) {
		return ea.fallback(aiCtx, FallbackReasonEarlyNoteIgnored, 0, now), nil
	}

	// At this point, we either have a breached SLA, OR a critical note (Delivery Failure/Cancellation). We proceed to call AI.
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

// fallback runs the rule-based analyzer and wraps the result.
func (ea *ExceptionAnalyzer) fallback(aiCtx *models.AIContext, reason string, durationMs int, now time.Time) *AnalysisResult {
	return ea.fallbackWithRaw(aiCtx, reason, durationMs, "", now)
	// return &AnalysisResult{
	// 	ExceptionType:      "OTHER",
	// 	Severity:           "LOW",
	// 	LikelyReason:       "No specific exception detected by rule-based analysis",
	// 	InternalNextAction: "No action required. Monitor the order for further changes.",
	// 	ConfidenceScore:    1.0,
	// 	FallbackUsed:       true,
	// 	FallbackReason:     reason,
	// 	DurationMs:         durationMs,
	// 	// RawResponse:        rawResponse,
	// }
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
		OrderID:         aiCtx.OrderID,
		CurrentStatus:   string(aiCtx.CurrentStatus),
		TotalAmount:     aiCtx.TotalAmount,
		CustomerName:    aiCtx.CustomerName,
		ShippingAddress: aiCtx.ShippingAddress,
		CreatedAt:       aiCtx.CreatedAt.Format(time.RFC3339),
		ErrorMessage:    notes,
		EventHistory:    eventHistory,
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
	case FallbackReasonNoDriverNote:
		return "No driver note provided, skipping AI analysis to save cost"
	case FallbackReasonEarlyNoteIgnored:
		return "Driver note ignored as SLA is not breached and note is not critical"
	default:
		return fmt.Sprintf("Unknown fallback reason: %s", reason)
	}
}
