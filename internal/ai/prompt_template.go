package ai

import (
	_ "embed"
	"fmt"
	"strings"

	"main/internal/dto"
)

// PromptTemplateVersion tracks the current version of the exception analysis prompt.
// Increment this when the prompt structure or instructions change.
const PromptTemplateVersion = "1.4.0"

const (
	MaxEventTimelineEntries = 50
	MaxDriverNotesLength    = 500
	MaxStringFieldLength    = 200
)

//go:embed knowledge/prompt_exception.md
var basePromptException string

//go:embed knowledge/prompt_draft.md
var basePromptDraft string

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
func BuildExceptionAnalysisPrompt(ctx ExceptionPromptContext, knowledge []KnowledgeEntry) string {
	// 1. Build context string
	var ctxSb strings.Builder
	ctxSb.WriteString(fmt.Sprintf("Order ID: %d\n", ctx.OrderID))
	ctxSb.WriteString(fmt.Sprintf("Current Status: %s\n", ctx.CurrentStatus))
	ctxSb.WriteString(fmt.Sprintf("Total Amount: %d VND\n", ctx.TotalAmount))
	ctxSb.WriteString(fmt.Sprintf("Customer Name: %s\n", ctx.CustomerName))
	ctxSb.WriteString(fmt.Sprintf("Shipping Address: %s\n", ctx.ShippingAddress))
	ctxSb.WriteString(fmt.Sprintf("Order Created At: %s\n", ctx.CreatedAt))
	ctxSb.WriteString(fmt.Sprintf("Analysis Timestamp (now): %s\n", ctx.AnalyzedAt))
	ctxSb.WriteString("\nEvent Timeline (chronological order):\n")

	if len(ctx.EventTimeline) == 0 {
		ctxSb.WriteString("  (no events recorded)\n")
	} else {
		for i, e := range ctx.EventTimeline {
			ctxSb.WriteString(fmt.Sprintf("  %d. [%s] %s → %s (by: %s)\n",
				i+1, e.EventAt, e.FromStatus, e.ToStatus, e.UpdatedBy))
		}
	}
	ctxSb.WriteString("\n")

	if ctx.DriverNotes != "" {
		ctxSb.WriteString(fmt.Sprintf("Driver Note: %s\n", ctx.DriverNotes))
	}

	// 2. Build knowledge base string
	if len(knowledge) == 0 {
		knowledge = ClassifyDriverNote("")
	}
	var kbSb strings.Builder
	for _, kb := range knowledge {
		kbSb.WriteString("--- ")
		kbSb.WriteString(kb.Title)
		kbSb.WriteString(" ---\n")
		kbSb.WriteString(kb.Body)
		kbSb.WriteString("\n\n")
	}

	// 3. Inject into base markdown template
	return fmt.Sprintf(basePromptException, ctxSb.String(), strings.TrimSpace(kbSb.String()))
}

func SanitizeCustomerUpdateDraftInput(input *dto.CustomerUpdateDraftInput) {
	if input.CustomerName != "" {
		input.BaselineDraft = strings.ReplaceAll(input.BaselineDraft, input.CustomerName, "[REDACTED_CUSTOMER_NAME]")
		input.CustomerName = "[REDACTED_CUSTOMER_NAME]"
	}
	if input.ShippingAddress != "" {
		input.BaselineDraft = strings.ReplaceAll(input.BaselineDraft, input.ShippingAddress, "[REDACTED_SHIPPING_ADDRESS]")
		input.ShippingAddress = "[REDACTED_SHIPPING_ADDRESS]"
	}
}

// BuildCustomerUpdateDraftPrompt constructs a prompt for drafting a customer update message.
func BuildCustomerUpdateDraftPrompt(input dto.CustomerUpdateDraftInput) string {
	// 1. Build context string
	var ctxSb strings.Builder
	ctxSb.WriteString(fmt.Sprintf("Order ID: %d\n", input.OrderID))
	ctxSb.WriteString(fmt.Sprintf("Customer Name: %s\n", input.CustomerName))
	ctxSb.WriteString(fmt.Sprintf("Shipping Address: %s\n", input.ShippingAddress))
	ctxSb.WriteString(fmt.Sprintf("Current Status: %s\n", input.CurrentStatus))
	ctxSb.WriteString(fmt.Sprintf("Exception Type: %s\n", input.ExceptionType))
	ctxSb.WriteString(fmt.Sprintf("Likely Reason: %s\n", input.LikelyReason))
	ctxSb.WriteString(fmt.Sprintf("Requested Tone: %s\n", input.Tone))
	ctxSb.WriteString(fmt.Sprintf("Communication Channel: %s\n", input.Channel))

	// 2. Inject into base markdown template
	return fmt.Sprintf(basePromptDraft, ctxSb.String(), input.BaselineDraft)
}
