package ai

import (
	_ "embed"
	"fmt"
	"strings"

	"main/internal/dto"
)

const PromptTemplateVersion = "1.0.1"

const (
	MaxEventTimelineEntries = 50
	MaxDriverNotesLength    = 500
	MaxStringFieldLength    = 200
)

//go:embed prompts/system_instruction.md
var systemInstructionMD string

//go:embed prompts/task_exception_analysis.md
var taskExceptionMD string

//go:embed prompts/task_draft_update.md
var taskDraftMD string

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

func BuildExceptionAnalysisPrompt(ctx ExceptionPromptContext) string {
	var sb strings.Builder

	// 1. System Instruction
	sb.WriteString(systemInstructionMD)
	sb.WriteString("\n\n")

	// 2. Prepare dynamic values
	timelineStr := ""
	if len(ctx.EventTimeline) == 0 {
		timelineStr = "  (no events recorded)\n"
	} else {
		for i, e := range ctx.EventTimeline {
			timelineStr += fmt.Sprintf("  %d. [%s] %s → %s (by: %s)\n",
				i+1, e.EventAt, e.FromStatus, e.ToStatus, e.UpdatedBy)
		}
	}

	kbSnippet := ""
	if ctx.DriverNotes != "" {
		kbSnippet = ClassifyDriverNote(ctx.DriverNotes).KBSnippet
	}

	// 3. Inject variables into the Task template
	prompt := taskExceptionMD
	prompt = strings.ReplaceAll(prompt, "{{KNOWLEDGE_BASE_INJECTION}}", kbSnippet)
	prompt = strings.ReplaceAll(prompt, "{{ORDER_ID}}", fmt.Sprintf("%d", ctx.OrderID))
	prompt = strings.ReplaceAll(prompt, "{{CURRENT_STATUS}}", ctx.CurrentStatus)
	prompt = strings.ReplaceAll(prompt, "{{TOTAL_AMOUNT}}", fmt.Sprintf("%d", ctx.TotalAmount))
	prompt = strings.ReplaceAll(prompt, "{{CREATED_AT}}", ctx.CreatedAt)
	prompt = strings.ReplaceAll(prompt, "{{ANALYZED_AT}}", ctx.AnalyzedAt)
	prompt = strings.ReplaceAll(prompt, "{{EVENT_TIMELINE}}", timelineStr)
	prompt = strings.ReplaceAll(prompt, "{{DRIVER_NOTES}}", ctx.DriverNotes)

	sb.WriteString(prompt)
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

	// 1. System Instruction
	sb.WriteString(systemInstructionMD)
	sb.WriteString("\n\n")

	// 2. Inject variables into the Task template
	prompt := taskDraftMD
	prompt = strings.ReplaceAll(prompt, "{{ORDER_ID}}", fmt.Sprintf("%d", input.OrderID))
	prompt = strings.ReplaceAll(prompt, "{{CURRENT_STATUS}}", input.CurrentStatus)
	prompt = strings.ReplaceAll(prompt, "{{EXCEPTION_TYPE}}", input.ExceptionType)
	prompt = strings.ReplaceAll(prompt, "{{LIKELY_REASON}}", input.LikelyReason)
	prompt = strings.ReplaceAll(prompt, "{{TONE}}", input.Tone)
	prompt = strings.ReplaceAll(prompt, "{{CHANNEL}}", input.Channel)

	sb.WriteString(prompt)
	return sb.String()
}
