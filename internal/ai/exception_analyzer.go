package ai

import (
	"context"
	"main/internal/dto"
	"main/internal/models"
	"strings"
	"time"
)

// Fallback reason constants used in audit logging.
const (
	FallbackReasonDisabled           = "ai_disabled"
	FallbackReasonNoDriverNote       = "no_driver_note"
	FallbackReasonEarlyNoteIgnored   = "early_note_ignored_to_save_cost"
	FallbackReasonTemplateSufficient = "template_sufficient"
	FallbackReasonTimeout            = "ai_timeout"
	FallbackReasonConnectionError    = "ai_connection_error"
	FallbackReasonInvalidResponse    = "ai_invalid_response"
	FallbackReasonLowConfidence      = "ai_confidence_below_threshold"
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

	// ── Step 1: Rule-based engine
	ruleResult := AnalyzeByRules(aiCtx, now)

	// ── Step 2: Check for driver notes
	hasDriverNote := hasAnyDriverNote(aiCtx) || strings.TrimSpace(notes) != ""

	// ── Step 3: Skip AI when disabled or no driver note
	if !ea.config.AIEnabled || !hasDriverNote {
		reason := FallbackReasonDisabled
		if ea.config.AIEnabled && !hasDriverNote {
			reason = FallbackReasonNoDriverNote
		}
		return ruleResultToAnalysis(ruleResult, reason, 0, ""), nil
	}

	// ── Step 3b: Skip AI if order is completely healthy (SLA not breached)
	if ruleResult == nil {
		return ea.fallbackWithRaw(aiCtx, FallbackReasonEarlyNoteIgnored, 0, "", now), nil
	}

	// ── Step 4: Call AI
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

func hasAnyDriverNote(aiCtx *models.AIContext) bool {
	for _, e := range aiCtx.Events {
		if e.DriverNote != nil && strings.TrimSpace(*e.DriverNote) != "" {
			return true
		}
	}
	return false
}

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

// fallback runs the rule-based analyzer
func (ea *ExceptionAnalyzer) fallback(aiCtx *models.AIContext, reason string, durationMs int, now time.Time) *AnalysisResult {
	return ea.fallbackWithRaw(aiCtx, reason, durationMs, "", now)
}

// fallbackWithRaw
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
