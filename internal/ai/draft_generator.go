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
	// Step 1: Check if AI is disabled — fallback immediately.
	if !dg.config.AIEnabled {
		return dg.fallback(input, FallbackReasonDisabled, 0, ""), nil
	}

	// Step 2: Call AI adapter with per-request timeout.
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
	message := buildFallbackDraftMessage(input)
	return &DraftResult{
		CustomerUpdateDraft: message,
		ConfidenceScore:     1.0, // Template is deterministic — confidence is certain
		FallbackUsed:        true,
		FallbackReason:      reason,
		DurationMs:          durationMs,
		RawResponse:         rawResponse,
	}
}

// buildFallbackDraftMessage generates a safe, generic customer update message
// using a template populated with order context. Adapts tone when possible.
func buildFallbackDraftMessage(input dto.CustomerUpdateDraftInput) string {
	customerName := input.CustomerName
	if customerName == "" {
		customerName = "Valued Customer"
	}

	switch strings.ToLower(input.Tone) {
	case "apologetic":
		return fmt.Sprintf(
			"Dear %s, we sincerely apologize for the inconvenience with your order #%d. "+
				"We are currently looking into the issue and will keep you updated as soon as possible. "+
				"Thank you for your patience.",
			customerName, input.OrderID,
		)
	case "informative":
		return fmt.Sprintf(
			"Hello %s, this is an update regarding your order #%d. "+
				"We have identified an issue and our team is actively working to resolve it. "+
				"We will notify you once the matter has been addressed.",
			customerName, input.OrderID,
		)
	case "proactive":
		return fmt.Sprintf(
			"Hi %s! We wanted to proactively reach out about your order #%d. "+
				"Our team has spotted an issue and is already working on a resolution. "+
				"We appreciate your patience and will update you shortly.",
			customerName, input.OrderID,
		)
	default: // neutral
		return fmt.Sprintf(
			"Hello %s, we are writing to inform you about your order #%d. "+
				"Our team is currently reviewing the situation and will provide an update as soon as possible. "+
				"We apologize for any inconvenience this may cause.",
			customerName, input.OrderID,
		)
	}
}

// FormatDraftFallbackReason returns a human-readable description of the fallback reason.
func FormatDraftFallbackReason(reason string) string {
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
