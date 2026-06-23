package ai

import (
	"fmt"
	"strings"

	"main/internal/dto"
)

// PromptTemplateVersion tracks the current version of the exception analysis prompt.
// Increment this when the prompt structure or instructions change.
const PromptTemplateVersion = "1.3.0"

const (
	MaxEventTimelineEntries = 50

	MaxDriverNotesLength = 500

	MaxStringFieldLength = 200
)

type ExceptionPromptContext struct {
	OrderID         int64
	CurrentStatus   string
	TotalAmount     int64
	CustomerName    string
	ShippingAddress string
	CreatedAt       string
	AnalyzedAt      string
	EventTimeline   []EventTimelineEntry
	DriverNotes     string
}

type EventTimelineEntry struct {
	FromStatus string
	ToStatus   string
	UpdatedBy  string
	EventAt    string
}

func SanitizePromptContext(ctx *ExceptionPromptContext) {
	if ctx.CustomerName != "" {
		ctx.CustomerName = "[REDACTED_CUSTOMER_NAME]"
	}
	if ctx.ShippingAddress != "" {
		ctx.ShippingAddress = "[REDACTED_SHIPPING_ADDRESS]"
	}

	if len(ctx.EventTimeline) > MaxEventTimelineEntries {
		ctx.EventTimeline = ctx.EventTimeline[len(ctx.EventTimeline)-MaxEventTimelineEntries:]
	}

	if len(ctx.DriverNotes) > MaxDriverNotesLength {
		ctx.DriverNotes = ctx.DriverNotes[:MaxDriverNotesLength] + "...[truncated]"
	}
}

// BuildExceptionAnalysisPrompt constructs a focused prompt for the AI.
//
// knowledge is the set of KnowledgeEntry items selected by ClassifyDriverNote;
// only the relevant domain rules are injected, keeping the prompt lean.
// Pass nil or an empty slice to fall back to the baseline state-machine KB only.
func BuildExceptionAnalysisPrompt(ctx ExceptionPromptContext, knowledge []KnowledgeEntry) string {
	var sb strings.Builder

	// ── [SYSTEM] ──────────────────────────────────────────────────────────────
	sb.WriteString("[SYSTEM]\n")
	sb.WriteString("You are an Order Exception Analyst for a fulfillment tracking system.\n")
	sb.WriteString("Your role is to analyze order data, event history, and operator notes to identify exceptions.\n")
	sb.WriteString("You must determine the exception type, severity, likely root cause, and recommend the next internal action.\n")
	sb.WriteString("Use ONLY the rules provided in [KNOWLEDGE BASE] to assign severity — do not use intuition or guesswork.\n")
	sb.WriteString("You MUST respond ONLY with a single valid JSON object. No explanations, no markdown, no extra text.\n\n")

	// ── [CONTEXT] ─────────────────────────────────────────────────────────────
	sb.WriteString("[CONTEXT]\n")
	sb.WriteString(fmt.Sprintf("Order ID: %d\n", ctx.OrderID))
	sb.WriteString(fmt.Sprintf("Current Status: %s\n", ctx.CurrentStatus))
	sb.WriteString(fmt.Sprintf("Total Amount: %d VND\n", ctx.TotalAmount))
	sb.WriteString(fmt.Sprintf("Customer Name: %s\n", ctx.CustomerName))
	sb.WriteString(fmt.Sprintf("Shipping Address: %s\n", ctx.ShippingAddress))
	sb.WriteString(fmt.Sprintf("Order Created At: %s\n", ctx.CreatedAt))
	sb.WriteString(fmt.Sprintf("Analysis Timestamp (now): %s\n", ctx.AnalyzedAt))
	sb.WriteString("\n")

	// Event timeline
	sb.WriteString("Event Timeline (chronological order):\n")
	if len(ctx.EventTimeline) == 0 {
		sb.WriteString("  (no events recorded)\n")
	} else {
		for i, e := range ctx.EventTimeline {
			sb.WriteString(fmt.Sprintf("  %d. [%s] %s → %s (by: %s)\n",
				i+1, e.EventAt, e.FromStatus, e.ToStatus, e.UpdatedBy))
		}
	}
	sb.WriteString("\n")

	if ctx.DriverNotes != "" {
		sb.WriteString(fmt.Sprintf("Driver Note: %s\n\n", ctx.DriverNotes))
	}

	// ── [KNOWLEDGE BASE] — injected per driver-note classification ────────────
	if len(knowledge) == 0 {
		knowledge = []KnowledgeEntry{kbStateMachine}
	}
	sb.WriteString("[KNOWLEDGE BASE]\n")
	sb.WriteString("Apply the following domain rules when classifying and rating the exception:\n\n")
	for _, kb := range knowledge {
		sb.WriteString("--- ")
		sb.WriteString(kb.Title)
		sb.WriteString(" ---\n")
		sb.WriteString(kb.Body)
		sb.WriteString("\n\n")
	}

	// ── [TASK] ────────────────────────────────────────────────────────────────
	sb.WriteString("[TASK]\n")
	sb.WriteString("Analyze the order context and the driver note above.\n")
	sb.WriteString("Identify the exception type and assign severity using ONLY the [KNOWLEDGE BASE] rules.\n")
	sb.WriteString("Focus on anomalies mentioned in the driver note as the primary signal.\n\n")

	// ── [OUTPUT FORMAT] ───────────────────────────────────────────────────────
	sb.WriteString("[OUTPUT FORMAT]\n")
	sb.WriteString("Respond with EXACTLY this JSON structure (no additional fields, no wrapping):\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"exception_type\": \"<string: one of INVALID_TRANSITION | STUCK_ORDER | SKIPPED_STATUS | DUPLICATE_EVENT | DELIVERY_FAILURE | CANCELLATION_ANOMALY | REFUND_ANOMALY | OTHER>\",\n")
	sb.WriteString("  \"severity\": \"<string: one of LOW | MEDIUM | HIGH | CRITICAL>\",\n")
	sb.WriteString("  \"likely_reason\": \"<string: concise root cause explanation in English, max 200 chars>\",\n")
	sb.WriteString("  \"internal_next_action\": \"<string: recommended internal action for the fulfillment team, max 200 chars>\",\n")
	sb.WriteString("  \"should_alert\": <boolean: true if the exception warrants an alert, false otherwise>,\n")
	sb.WriteString("  \"confidence_score\": <float: 0.0 to 1.0, your confidence in this analysis>\n")
	sb.WriteString("}\n\n")

	// ── [CONSTRAINTS] ─────────────────────────────────────────────────────────
	sb.WriteString("[CONSTRAINTS]\n")
	sb.WriteString("1. You MUST NOT suggest or imply any automatic order status changes.\n")
	sb.WriteString("2. You MUST NOT suggest or trigger any refund actions.\n")
	sb.WriteString("3. You MUST NOT suggest sending any messages or notifications to customers.\n")
	sb.WriteString("4. Your role is ANALYSIS ONLY — observe, diagnose, and recommend internal actions.\n")
	sb.WriteString("5. If you cannot determine the exception with reasonable confidence, set confidence_score below 0.5.\n")
	sb.WriteString("6. Do NOT output anything other than the JSON object. No markdown fences, no explanations.\n")

	return sb.String()
}

func SanitizeCustomerUpdateDraftInput(input *dto.CustomerUpdateDraftInput) {
	if input.CustomerName != "" {
		input.CustomerName = "[REDACTED_CUSTOMER_NAME]"
	}
	if input.ShippingAddress != "" {
		input.ShippingAddress = "[REDACTED_SHIPPING_ADDRESS]"
	}
}

func BuildCustomerUpdateDraftPrompt(input dto.CustomerUpdateDraftInput) string {
	var sb strings.Builder

	// ── [SYSTEM] ──────────────────────────────────────────────────────────────
	sb.WriteString("[SYSTEM]\n")
	sb.WriteString("You are a Customer Support Agent for a fulfillment tracking system.\n")
	sb.WriteString("Your job is to draft a customer-facing update message regarding an order exception.\n")
	sb.WriteString("You MUST avoid making unsupported promises (like guaranteed delivery times) and MUST NOT leak internal technical details.\n")
	sb.WriteString("You MUST respond ONLY with a single valid JSON object. No explanations, no markdown, no extra text.\n\n")

	// ── [CONTEXT] ─────────────────────────────────────────────────────────────
	sb.WriteString("[CONTEXT]\n")
	sb.WriteString(fmt.Sprintf("Order ID: %d\n", input.OrderID))
	sb.WriteString(fmt.Sprintf("Customer Name: %s\n", input.CustomerName))
	sb.WriteString(fmt.Sprintf("Shipping Address: %s\n", input.ShippingAddress))
	sb.WriteString(fmt.Sprintf("Current Status: %s\n", input.CurrentStatus))
	sb.WriteString(fmt.Sprintf("Exception Type: %s\n", input.ExceptionType))
	sb.WriteString(fmt.Sprintf("Likely Reason: %s\n", input.LikelyReason))
	sb.WriteString(fmt.Sprintf("Requested Tone: %s\n", input.Tone))
	sb.WriteString(fmt.Sprintf("Communication Channel: %s\n", input.Channel))
	sb.WriteString("\n")

	// ── [TASK] ────────────────────────────────────────────────────────────────
	sb.WriteString("[TASK]\n")
	sb.WriteString("Draft a message to the customer explaining the situation politely, using the requested tone and appropriate length for the channel.\n")
	sb.WriteString("Reassure them that we are handling it, but do not promise refunds or exact resolution times unless explicitly supported by standard policy.\n\n")

	// ── [OUTPUT FORMAT] ───────────────────────────────────────────────────────
	sb.WriteString("[OUTPUT FORMAT]\n")
	sb.WriteString("Respond with EXACTLY this JSON structure:\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"customer_update_draft\": \"<string: the drafted message for the customer>\",\n")
	sb.WriteString("  \"confidence_score\": <float: 0.0 to 1.0, your confidence in the appropriateness of this draft>\n")
	sb.WriteString("}\n\n")

	// ── [CONSTRAINTS] ─────────────────────────────────────────────────────────
	sb.WriteString("[CONSTRAINTS]\n")
	sb.WriteString("1. NO internal technical jargon.\n")
	sb.WriteString("2. NO false promises.\n")
	sb.WriteString("3. DO NOT output anything other than the JSON object.\n")

	return sb.String()
}
