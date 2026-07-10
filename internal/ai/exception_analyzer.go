package ai

import (
	"context"
	"main/internal/dto"
	"main/internal/models"
	"strings"
	"time"
	"unicode"
)

const (
	FallbackReasonDisabled            = "ai_disabled"
	FallbackReasonNoDriverNote        = "no_driver_note"
	FallbackReasonEarlyNoteIgnored    = "early_note_ignored_to_save_cost"
	FallbackReasonTemplateSufficient  = "template_sufficient"
	FallbackReasonTimeout             = "ai_timeout"
	FallbackReasonConnectionError     = "ai_connection_error"
	FallbackReasonInvalidResponse     = "ai_invalid_response"
	FallbackReasonLowConfidence       = "ai_confidence_below_threshold"
	FallbackReasonStructuralRuleMatch = "structural_rule_match"
	FallbackReasonNoteNotActionable   = "note_not_actionable"
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

func (ea *ExceptionAnalyzer) ReloadKnowledge(ctx context.Context) error {
	return ea.adapter.ReloadKnowledge(ctx)
}

func (ea *ExceptionAnalyzer) Analyze(ctx context.Context, aiCtx *models.AIContext) (*AnalysisResult, error) {
	now := time.Now()

	ruleResult := AnalyzeByRules(aiCtx, now)

	hasDriverNote := hasAnyDriverNote(aiCtx)

	isTerminalAnomaly := aiCtx.CurrentStatus == models.ORDER_STATUS_CANCELLED ||
		aiCtx.CurrentStatus == models.ORDER_STATUS_REFUNDED

	if !ea.config.AIEnabled {
		return ruleResultToAnalysis(ruleResult, FallbackReasonDisabled, 0, ""), nil
	}

	if !hasDriverNote && !isTerminalAnomaly {
		return ruleResultToAnalysis(ruleResult, FallbackReasonNoDriverNote, 0, ""), nil
	}

	if isTerminalAnomaly && !hasDriverNote {
		return ruleResultToAnalysis(ruleResult, FallbackReasonNoDriverNote, 0, ""), nil
	}

	if ruleResult != nil && isStructuralException(ruleResult.ExceptionType) {
		return ruleResultToAnalysis(ruleResult, FallbackReasonStructuralRuleMatch, 0, ""), nil
	}

	if hasDriverNote {
		var latestNote string
		for i := len(aiCtx.Events) - 1; i >= 0; i-- {
			e := aiCtx.Events[i]
			if e.DriverNote != nil && strings.TrimSpace(*e.DriverNote) != "" {
				latestNote = strings.TrimSpace(*e.DriverNote)
				break
			}
		}
		if !isNoteActionable(latestNote) {
			return ruleResultToAnalysis(ruleResult, FallbackReasonNoteNotActionable, 0, ""), nil
		}
	}

	input := buildExceptionInput(aiCtx)

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

func (ea *ExceptionAnalyzer) fallback(aiCtx *models.AIContext, reason string, durationMs int, now time.Time) *AnalysisResult {
	return ea.fallbackWithRaw(aiCtx, reason, durationMs, "", now)
}

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

func buildExceptionInput(aiCtx *models.AIContext) dto.ExceptionInput {
	eventHistory := make([]dto.EventRecord, 0, len(aiCtx.Events))
	for _, e := range aiCtx.Events {
		eventHistory = append(eventHistory, dto.EventRecord{
			FromStatus: string(e.PreviousStatus),
			ToStatus:   string(e.NewStatus),
			EventAt:    e.EventAt.Format(time.RFC3339),
			UpdatedBy:  e.UpdatedBy,
		})
	}

	var driverNotes string
	for i := len(aiCtx.Events) - 1; i >= 0; i-- {
		e := aiCtx.Events[i]
		if e.DriverNote != nil && strings.TrimSpace(*e.DriverNote) != "" {
			driverNotes = strings.TrimSpace(*e.DriverNote)
			break
		}
	}

	return dto.ExceptionInput{
		OrderID:         aiCtx.OrderID,
		CurrentStatus:   string(aiCtx.CurrentStatus),
		TotalAmount:     aiCtx.TotalAmount,
		CustomerName:    aiCtx.CustomerName,
		ShippingAddress: aiCtx.ShippingAddress,
		CreatedAt:       aiCtx.CreatedAt.Format(time.RFC3339),
		DriverNotes:     driverNotes,
		EventHistory:    eventHistory,
	}
}

func isStructuralException(exceptionType string) bool {
	switch exceptionType {
	case "INVALID_TRANSITION",
		"SKIPPED_STATUS",
		"DUPLICATE_EVENT",
		"CANCELLATION_ANOMALY",
		"REFUND_ANOMALY",
		"NONE":
		return true
	}
	return false
}

func isNoteActionable(note string) bool {
	s := strings.TrimSpace(note)

	if len([]rune(s)) < 10 {
		return false
	}
	letterCount := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			letterCount++
		}
	}
	if float64(letterCount)/float64(len([]rune(s))) < 0.4 {
		return false
	}

	if len(strings.Fields(s)) < 3 {
		return false
	}

	return true
}
