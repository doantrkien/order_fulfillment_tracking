package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"main/internal/dto"
)

// DraftResult is the unified output produced by DraftGenerator.
// Both AI and template-based fallback paths produce this same type.
type DraftResult struct {
	CustomerUpdateDraft string
	ConfidenceScore     float64
	FallbackUsed        bool
	FallbackReason      string
	DurationMs          int
	RawResponse         string // Raw AI response (empty when fallback)
}

// DraftGeneratorConfig holds runtime configuration for the generator.
type DraftGeneratorConfig struct {
	AIEnabled bool
	AITimeout time.Duration
}

// DraftGenerator orchestrates AI customer-update draft generation with
// automatic fallback to a template-based message when AI is unavailable.
//
// It mirrors the ExceptionAnalyzer pattern:
//  1. Check if AI is enabled → if not, fallback immediately
//  2. Call AI adapter with timeout context
//  3. Handle transport errors → fallback with classified reason
//  4. Parse and validate JSON output → fallback if malformed
//  5. Check confidence threshold → fallback if below minimum
//  6. Return the AI-generated draft
//
// This method never returns an error for AI failures — it always falls back
// to a safe template-based draft.
type DraftGenerator struct {
	adapter AIAdapter
	config  DraftGeneratorConfig
}

// NewDraftGenerator creates a DraftGenerator wired to the given adapter and config.
func NewDraftGenerator(adapter AIAdapter, config DraftGeneratorConfig) *DraftGenerator {
	return &DraftGenerator{
		adapter: adapter,
		config:  config,
	}
}

// Generate is the main entry point. See DraftGenerator doc for the full flow.
func (dg *DraftGenerator) Generate(ctx context.Context, input dto.CustomerUpdateDraftInput) (*DraftResult, error) {
	// Step 1: Pre-compute the baseline draft using the static template.
	// We inject this into the input so the AI can use it as a reference if called.
	input.BaselineDraft = buildFallbackDraftMessage(input)

	// Step 2: Decide whether we actually need to call the AI.
	// If AI is disabled or the static template is sufficient, return the static template immediately.
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

	// Step 3: Call AI adapter with per-request timeout.
	start := time.Now()
	timeoutCtx, cancel := context.WithTimeout(ctx, dg.config.AITimeout)
	defer cancel()

	rawText, err := dg.adapter.DraftCustomerUpdate(timeoutCtx, input)
	durationMs := int(time.Since(start).Milliseconds())

	// Step 3: Handle transport / timeout errors.
	if err != nil {
		reason := ClassifyError(err)
		return dg.fallback(input, reason, durationMs, ""), nil
	}

	// Step 4: Strip markdown fences then parse JSON.
	cleaned := stripDraftFences(rawText)
	var output dto.CustomerUpdateDraftOutput
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return dg.fallback(input, FallbackReasonInvalidResponse, durationMs, rawText), nil
	}

	// Step 5: Validate the draft content is non-empty.
	if strings.TrimSpace(output.CustomerUpdateDraft) == "" {
		return dg.fallback(input, FallbackReasonInvalidResponse, durationMs, rawText), nil
	}

	// Step 6: Check confidence threshold.
	if output.ConfidenceScore < ConfidenceThreshold {
		return dg.fallback(input, FallbackReasonLowConfidence, durationMs, rawText), nil
	}

	// Step 7: AI succeeded — return the generated draft.
	return &DraftResult{
		CustomerUpdateDraft: output.CustomerUpdateDraft,
		ConfidenceScore:     output.ConfidenceScore,
		FallbackUsed:        false,
		DurationMs:          durationMs,
		RawResponse:         rawText,
	}, nil
}

// fallback generates a safe template-based draft message when AI is unavailable.
// It preserves the raw AI response (if any) for audit purposes.
func (dg *DraftGenerator) fallback(input dto.CustomerUpdateDraftInput, reason string, durationMs int, rawResponse string) *DraftResult {
	message := input.BaselineDraft
	if message == "" {
		message = buildFallbackDraftMessage(input)
	}
	return &DraftResult{
		CustomerUpdateDraft: message,
		ConfidenceScore:     1.0, // Template is deterministic — confidence is certain
		FallbackUsed:        true,
		FallbackReason:      reason,
		DurationMs:          durationMs,
		RawResponse:         rawResponse,
	}
}

// shouldCallAI determines whether the AI is necessary for the given input.
// We only call the AI when the static template is insufficient.
func shouldCallAI(input dto.CustomerUpdateDraftInput) bool {
	// 1. If exception type is OTHER, we need AI to explain the LikelyReason.
	if input.ExceptionType == "OTHER" {
		return true
	}

	// 2. If a specific tone is requested (other than neutral/informative), we need AI to rewrite it.
	tone := strings.ToLower(strings.TrimSpace(input.Tone))
	if tone != "" && tone != "neutral" && tone != "informative" {
		return true
	}

	// 3. If the channel is SMS, we need AI to shorten the message.
	if strings.ToLower(strings.TrimSpace(input.Channel)) == "sms" {
		return true
	}

	// For standard exceptions, neutral tone, and standard channels (email), the template is sufficient.
	return false
}

// buildFallbackDraftMessage generates a safe, generic customer update message
// using a template populated with order context. Adapts tone when possible.
//
// In Phase 3 (AI integration), we pivot to static baseline templates in Vietnamese
// which leverage server-side/client-side placeholders [REDACTED_CUSTOMER_NAME] and
// [REDACTED_SHIPPING_ADDRESS] to support PII masking rules.
func buildFallbackDraftMessage(input dto.CustomerUpdateDraftInput) string {
	// 1. Retrieve the appropriate Vietnamese fallback template from our map using ExceptionType.
	// 2. We do not do inline replacement (hydration) of [REDACTED_CUSTOMER_NAME] and
	//    [REDACTED_SHIPPING_ADDRESS] here, as they serve as PII-safe placeholders
	//    that must be preserved for client/server compliance contract.
	return GetFallbackTemplate(input.ExceptionType)
}

// FormatDraftFallbackReason returns a human-readable description of the fallback reason.
func FormatDraftFallbackReason(reason string) string {
	switch reason {
	case FallbackReasonDisabled:
		return "AI feature is disabled via configuration"
	case FallbackReasonTemplateSufficient:
		return "Static template is sufficient, skipping AI draft generation"
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

// stripDraftFences removes ```json ... ``` wrappers that some LLMs add.
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
