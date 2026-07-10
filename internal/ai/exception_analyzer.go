package ai

import (
	"context"

	"main/constant"
	"main/errs"
	dto_ai "main/internal/dto/ai"
	"main/internal/models"
	"main/utils/helpers"
	"main/utils/validates"
	"strings"
	"time"
)

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

func ruleResultToAnalysis(ruleResult *RuleBasedResult, reason string, durationMs int, rawResponse string) *dto_ai.AIAnalysisResult {
	if ruleResult == nil {
		return &dto_ai.AIAnalysisResult{
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
	return &dto_ai.AIAnalysisResult{
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

func (ea *ExceptionAnalyzer) ReloadKnowledge(ctx context.Context) error {
	return ea.adapter.ReloadKnowledge(ctx)
}

func (ea *ExceptionAnalyzer) Analyze(ctx context.Context, aiCtx *models.AIContext) (*dto_ai.AIAnalysisResult, error) {
	now := time.Now()

	ruleResult := AnalyzeByRules(aiCtx, now)

	hasDriverNote := validates.HasAnyDriverNote(aiCtx)

	isTerminalAnomaly := aiCtx.CurrentStatus == models.ORDER_STATUS_CANCELLED ||
		aiCtx.CurrentStatus == models.ORDER_STATUS_REFUNDED

	if !ea.config.AIEnabled {
		return ruleResultToAnalysis(ruleResult, constant.FallbackReasonDisabled, 0, ""), nil
	}

	if !hasDriverNote && !isTerminalAnomaly {
		return ruleResultToAnalysis(ruleResult, constant.FallbackReasonNoDriverNote, 0, ""), nil
	}

	if isTerminalAnomaly && !hasDriverNote {
		return ruleResultToAnalysis(ruleResult, constant.FallbackReasonNoDriverNote, 0, ""), nil
	}

	if ruleResult != nil && validates.IsStructuralException(ruleResult.ExceptionType) {
		return ruleResultToAnalysis(ruleResult, constant.FallbackReasonStructuralRuleMatch, 0, ""), nil
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
		if !validates.IsNoteActionable(latestNote) {
			return ruleResultToAnalysis(ruleResult, constant.FallbackReasonNoteNotActionable, 0, ""), nil
		}
	}

	input := helpers.BuildExceptionInput(aiCtx)

	start := time.Now()
	aiCtxTimeout, cancel := context.WithTimeout(ctx, ea.config.AITimeout)
	defer cancel()

	aiOutput, rawText, err := ea.adapter.AnalyzeException(aiCtxTimeout, input)
	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		reason := errs.ClassifyError(err)
		return ea.fallbackWithRaw(aiCtx, reason, durationMs, rawText, now), nil
	}

	if validationErr := ParseAndValidateAIOutput(&aiOutput); validationErr != nil {
		return ea.fallbackWithRaw(aiCtx, constant.FallbackReasonInvalidResponse, durationMs, rawText, now), nil
	}

	if shouldFallback, _ := validates.ShouldFallback(&aiOutput); shouldFallback {
		return ea.fallbackWithRaw(aiCtx, constant.FallbackReasonLowConfidence, durationMs, rawText, now), nil
	}

	return &dto_ai.AIAnalysisResult{
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

func (ea *ExceptionAnalyzer) fallback(aiCtx *models.AIContext, reason string, durationMs int, now time.Time) *dto_ai.AIAnalysisResult {
	return ea.fallbackWithRaw(aiCtx, reason, durationMs, "", now)
}

func (ea *ExceptionAnalyzer) fallbackWithRaw(aiCtx *models.AIContext, reason string, durationMs int, rawResponse string, now time.Time) *dto_ai.AIAnalysisResult {
	ruleResult := AnalyzeByRules(aiCtx, now)

	if ruleResult == nil {
		return &dto_ai.AIAnalysisResult{
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

	return &dto_ai.AIAnalysisResult{
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
