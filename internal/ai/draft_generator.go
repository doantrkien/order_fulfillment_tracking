package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"main/constant"
	"main/errs"
	dto_ai "main/internal/dto/ai"
	"main/utils/helpers"
	"main/utils/validates"
)

const (
	ExceptionAlternativeDelivery = "ALTERNATIVE_DELIVERY"
	ExceptionOther               = "OTHER"

	ChannelSMS     = "sms"
	ToneApologetic = "apologetic"
	ToneProactive  = "proactive"
)

type DraftResult struct {
	CustomerUpdateDraft string
	ConfidenceScore     float64
	FallbackUsed        bool
	FallbackReason      string
	DurationMs          int
	RawResponse         string
}

type DraftGeneratorConfig struct {
	AIEnabled bool
	AITimeout time.Duration

	ScoreChannelSMS              int
	ScoreToneApologeticProactive int
	ScoreLongReason              int
	LikelyReasonLengthThreshold  int
	AIScoreThreshold             int
}

type DraftGenerator struct {
	adapter AIAdapter
	config  DraftGeneratorConfig
}

func NewDraftGenerator(adapter AIAdapter, config DraftGeneratorConfig) *DraftGenerator {
	return &DraftGenerator{
		adapter: adapter,
		config:  config,
	}
}

func (dg *DraftGenerator) Generate(ctx context.Context, input dto_ai.CustomerUpdateDraftInput) (*DraftResult, error) {

	input.BaselineDraft = buildFallbackDraftMessage(input)

	if !dg.config.AIEnabled || !dg.shouldCallAI(input) {
		reason := constant.FallbackReasonTemplateSufficient
		if !dg.config.AIEnabled {
			reason = constant.FallbackReasonDisabled
		}
		return &DraftResult{
			CustomerUpdateDraft: input.BaselineDraft,
			ConfidenceScore:     1.0,
			FallbackUsed:        true,
			FallbackReason:      reason,
			DurationMs:          0,
		}, nil
	}

	start := time.Now()
	timeoutCtx, cancel := context.WithTimeout(ctx, dg.config.AITimeout)
	defer cancel()

	rawText, err := dg.adapter.DraftCustomerUpdate(timeoutCtx, input)
	durationMs := int(time.Since(start).Milliseconds())
	if err != nil {
		reason := errs.ClassifyError(err)
		return dg.fallback(input, reason, durationMs, ""), nil
	}

	cleaned := helpers.StripFences(rawText)
	var output dto_ai.CustomerUpdateDraftOutput
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return dg.fallback(input, constant.FallbackReasonInvalidResponse, durationMs, rawText), nil
	}

	if strings.TrimSpace(output.CustomerUpdateDraft) == "" {
		return dg.fallback(input, constant.FallbackReasonInvalidResponse, durationMs, rawText), nil
	}

	if output.ConfidenceScore < validates.ConfidenceThreshold {
		return dg.fallback(input, constant.FallbackReasonLowConfidence, durationMs, rawText), nil
	}

	return &DraftResult{
		CustomerUpdateDraft: output.CustomerUpdateDraft,
		ConfidenceScore:     output.ConfidenceScore,
		FallbackUsed:        false,
		DurationMs:          durationMs,
		RawResponse:         rawText,
	}, nil
}

func (dg *DraftGenerator) fallback(input dto_ai.CustomerUpdateDraftInput, reason string, durationMs int, rawResponse string) *DraftResult {
	message := input.BaselineDraft
	if message == "" {
		message = buildFallbackDraftMessage(input)
	}
	return &DraftResult{
		CustomerUpdateDraft: message,
		ConfidenceScore:     1.0,
		FallbackUsed:        true,
		FallbackReason:      reason,
		DurationMs:          durationMs,
		RawResponse:         rawResponse,
	}
}

func (dg *DraftGenerator) shouldCallAI(input dto_ai.CustomerUpdateDraftInput) bool {
	channel := strings.ToLower(strings.TrimSpace(input.Channel))
	tone := strings.ToLower(strings.TrimSpace(input.Tone))

	if validates.IsInternalSystemError(input.ExceptionType) {
		fmt.Printf("[DEBUG][shouldCallAI] VETO: Internal System Error (%s)\n", input.ExceptionType)
		return false
	}
	if input.ExceptionType == ExceptionAlternativeDelivery {
		fmt.Printf("[DEBUG][shouldCallAI] VETO: Alternative Delivery\n")
		return false
	}

	if input.ExceptionType == ExceptionOther {
		fmt.Printf("[DEBUG][shouldCallAI] MUST: Exception Type is OTHER\n")
		return true
	}

	score := 0

	if channel == ChannelSMS {
		score += dg.config.ScoreChannelSMS
	}

	if tone == ToneApologetic || tone == ToneProactive {
		score += dg.config.ScoreToneApologeticProactive
	}
	if len(input.LikelyReason) > dg.config.LikelyReasonLengthThreshold {
		score += dg.config.ScoreLongReason
	}

	needAI := score >= dg.config.AIScoreThreshold

	fmt.Printf("[DEBUG][shouldCallAI] ExceptionType: %s | Tone: %q | Channel: %q | Score: %d/%d | NeedAI: %v\n",
		input.ExceptionType, tone, channel, score, dg.config.AIScoreThreshold, needAI)

	return needAI
}

func buildFallbackDraftMessage(input dto_ai.CustomerUpdateDraftInput) string {
	return GetFallbackTemplate(input.ExceptionType, input.CustomerName, input.ShippingAddress, input.CurrentStatus)
}
