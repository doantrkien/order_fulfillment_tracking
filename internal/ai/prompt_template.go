package ai

import (
	_ "embed"
	"fmt"
	"strings"

	"main/constant"
	dto_ai "main/internal/dto/ai"
)

//go:embed knowledge/prompt_exception.md
var basePromptException string

//go:embed knowledge/prompt_draft.md
var basePromptDraft string

func SanitizePromptContext(ctx *dto_ai.ExceptionPromptContext) {
	if ctx.CustomerName != "" {
		ctx.CustomerName = "[REDACTED_CUSTOMER_NAME]"
	}
	if ctx.ShippingAddress != "" {
		ctx.ShippingAddress = "[REDACTED_SHIPPING_ADDRESS]"
	}

	if len(ctx.EventTimeline) > constant.MaxEventTimelineEntries {
		ctx.EventTimeline = ctx.EventTimeline[len(ctx.EventTimeline)-constant.MaxEventTimelineEntries:]
	}

	if len(ctx.DriverNotes) > constant.MaxDriverNotesLength {
		ctx.DriverNotes = ctx.DriverNotes[:constant.MaxDriverNotesLength] + "...[truncated]"
	}
}

func BuildExceptionAnalysisPrompt(ctx dto_ai.ExceptionPromptContext, knowledge []KnowledgeEntry) string {
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

	if len(knowledge) == 0 {
		knowledge = GetKnowledgeBase("")
	}
	var kbSb strings.Builder
	for _, kb := range knowledge {
		kbSb.WriteString("--- ")
		kbSb.WriteString(kb.Title)
		kbSb.WriteString(" ---\n")
		kbSb.WriteString(kb.Body)
		kbSb.WriteString("\n\n")
	}

	return fmt.Sprintf(basePromptException, ctxSb.String(), strings.TrimSpace(kbSb.String()))
}

func SanitizeCustomerUpdateDraftInput(input *dto_ai.CustomerUpdateDraftInput) {
	if input.CustomerName != "" {
		input.CustomerName = "[REDACTED_CUSTOMER_NAME]"
	}
	if input.ShippingAddress != "" {
		input.ShippingAddress = "[REDACTED_SHIPPING_ADDRESS]"
	}
}

func BuildCustomerUpdateDraftPrompt(input dto_ai.CustomerUpdateDraftInput) string {
	var ctxSb strings.Builder
	ctxSb.WriteString(fmt.Sprintf("Order ID: %d\n", input.OrderID))
	ctxSb.WriteString(fmt.Sprintf("Customer Name: %s\n", input.CustomerName))
	ctxSb.WriteString(fmt.Sprintf("Shipping Address: %s\n", input.ShippingAddress))
	ctxSb.WriteString(fmt.Sprintf("Current Status: %s\n", input.CurrentStatus))
	ctxSb.WriteString(fmt.Sprintf("Exception Type: %s\n", input.ExceptionType))
	ctxSb.WriteString(fmt.Sprintf("Likely Reason: %s\n", input.LikelyReason))
	ctxSb.WriteString(fmt.Sprintf("Requested Tone: %s\n", input.Tone))
	ctxSb.WriteString(fmt.Sprintf("Communication Channel: %s\n", input.Channel))

	return fmt.Sprintf(basePromptDraft, ctxSb.String(), input.BaselineDraft)
}
