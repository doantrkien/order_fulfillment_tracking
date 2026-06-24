package ai

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"main/internal/dto"
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

func (dg *DraftGenerator) Generate(ctx context.Context, input dto.CustomerUpdateDraftInput) (*DraftResult, error) {

	input.BaselineDraft = buildFallbackDraftMessage(input)

	if !dg.config.AIEnabled || !shouldCallAI(input) {
		reason := FallbackReasonTemplateSufficient
		if !dg.config.AIEnabled {
			reason = FallbackReasonDisabled
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
		reason := ClassifyError(err)
		return dg.fallback(input, reason, durationMs, ""), nil
	}

	cleaned := stripDraftFences(rawText)
	var output dto.CustomerUpdateDraftOutput
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return dg.fallback(input, FallbackReasonInvalidResponse, durationMs, rawText), nil
	}

	if strings.TrimSpace(output.CustomerUpdateDraft) == "" {
		return dg.fallback(input, FallbackReasonInvalidResponse, durationMs, rawText), nil
	}

	if output.ConfidenceScore < ConfidenceThreshold {
		return dg.fallback(input, FallbackReasonLowConfidence, durationMs, rawText), nil
	}

	return &DraftResult{
		CustomerUpdateDraft: output.CustomerUpdateDraft,
		ConfidenceScore:     output.ConfidenceScore,
		FallbackUsed:        false,
		DurationMs:          durationMs,
		RawResponse:         rawText,
	}, nil
}

// fallback generates a safe template-based draft message when AI is unavailable.
func (dg *DraftGenerator) fallback(input dto.CustomerUpdateDraftInput, reason string, durationMs int, rawResponse string) *DraftResult {
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

// shouldCallAI determines whether the AI is necessary for the given input.
func shouldCallAI(input dto.CustomerUpdateDraftInput) bool {
	// 1. If exception type is OTHER, we need AI to explain the LikelyReason.
	if input.ExceptionType == "OTHER" {
		return true
	}

	// 2. If a specific tone is requested
	tone := strings.ToLower(strings.TrimSpace(input.Tone))
	if tone != "" && tone != "neutral" && tone != "informative" {
		return true
	}

	// 3. If the channel is SMS
	if strings.ToLower(strings.TrimSpace(input.Channel)) == "sms" {
		return true
	}

	return false
}

func buildFallbackDraftMessage(input dto.CustomerUpdateDraftInput) string {
	return GetFallbackTemplate(input.ExceptionType)
}

func stripDraftFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	return s
}
