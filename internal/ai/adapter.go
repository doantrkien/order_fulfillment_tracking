package ai

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"main/errs"
	"main/internal/dto"
	"main/pkg/aiclient"
)

type AIAdapter interface {
	AnalyzeException(ctx context.Context, input dto.ExceptionInput) (dto.ExceptionOutput, string, error)
	SummarizeReport(ctx context.Context, input dto.ExceptionOutput) (dto.ReportSummaryOutput, error)
	DraftCustomerUpdate(ctx context.Context, input dto.CustomerUpdateDraftInput) (string, error)
	Ping(ctx context.Context) error
}

type aiAdapter struct {
	client aiclient.AIClient
}

func NewAIAdapter(client aiclient.AIClient) *aiAdapter {
	return &aiAdapter{client: client}
}

func (g *aiAdapter) AnalyzeException(
	ctx context.Context,
	input dto.ExceptionInput,
) (dto.ExceptionOutput, string, error) {
	timeline := make([]EventTimelineEntry, 0, len(input.EventHistory))
	for _, e := range input.EventHistory {
		timeline = append(timeline, EventTimelineEntry{
			FromStatus: e.FromStatus,
			ToStatus:   e.ToStatus,
			UpdatedBy:  e.UpdatedBy,
			EventAt:    e.EventAt,
		})
	}

	promptCtx := ExceptionPromptContext{
		OrderID:         input.OrderID,
		CurrentStatus:   input.CurrentStatus,
		TotalAmount:     input.TotalAmount,
		CustomerName:    input.CustomerName,
		ShippingAddress: input.ShippingAddress,
		CreatedAt:       input.CreatedAt,
		DriverNotes:     input.ErrorMessage,
		EventTimeline:   timeline,
		AnalyzedAt:      time.Now().Format(time.RFC3339),
	}

	SanitizePromptContext(&promptCtx)

	// Select only the KB entries relevant to this driver note.
	// The AI receives focused, context-appropriate rules instead of the full domain dump.
	knowledge := ClassifyDriverNote(input.ErrorMessage)
	prompt := BuildExceptionAnalysisPrompt(promptCtx, knowledge)

	rawText, err := g.client.GenerateContent(ctx, prompt)
	if err != nil {
		return dto.ExceptionOutput{}, "", errs.ERR_GEMINI_GENERATE_CONTENT_FAILED
	}

	var output dto.ExceptionOutput
	cleaned := stripMarkdownFences(rawText)
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return dto.ExceptionOutput{}, rawText, errs.ERR_AI_RESPONSE_VALIDATION_FAILED
	}

	return output, rawText, nil
}

func (g *aiAdapter) DraftCustomerUpdate(
	ctx context.Context,
	input dto.CustomerUpdateDraftInput,
) (string, error) {
	SanitizeCustomerUpdateDraftInput(&input)
	prompt := BuildCustomerUpdateDraftPrompt(input)

	rawText, err := g.client.GenerateContent(ctx, prompt)
	if err != nil {
		return "", errs.ERR_GEMINI_GENERATE_CONTENT_FAILED
	}

	return rawText, nil
}

func (g *aiAdapter) SummarizeReport(
	ctx context.Context,
	input dto.ExceptionOutput,
) (dto.ReportSummaryOutput, error) {
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return dto.ReportSummaryOutput{}, errs.ERR_INTERNAL_SERVER
	}

	prompt := buildReportSummaryPrompt(string(inputJSON))

	rawText, err := g.client.GenerateContent(ctx, prompt)
	if err != nil {
		return dto.ReportSummaryOutput{}, errs.ERR_GEMINI_GENERATE_CONTENT_FAILED
	}

	var output dto.ReportSummaryOutput
	cleaned := stripMarkdownFences(rawText)
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return dto.ReportSummaryOutput{}, errs.ERR_AI_RESPONSE_VALIDATION_FAILED
	}

	return output, nil
}

func (g *aiAdapter) Ping(ctx context.Context) error {
	_, err := g.client.GenerateContent(ctx, "ping")
	if err != nil {
		return errs.ERR_AI_PING_FAILED
	}
	return nil
}

func buildReportSummaryPrompt(reportJSON string) string {
	var sb strings.Builder

	sb.WriteString("[SYSTEM]\n")
	sb.WriteString("You are a fulfillment analytics assistant.\n")
	sb.WriteString("Your job is to summarize an order exception report into a concise human-readable format.\n")
	sb.WriteString("You MUST respond ONLY with a single valid JSON object. No explanations, no markdown.\n\n")

	sb.WriteString("[INPUT]\n")
	sb.WriteString(reportJSON)
	sb.WriteString("\n\n")

	sb.WriteString("[OUTPUT FORMAT]\n")
	sb.WriteString("Respond with EXACTLY this JSON structure:\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"summary\": \"<string: 1-3 sentence overall summary>\",\n")
	sb.WriteString("  \"highlights\": [\"<string>\", ...],\n")
	sb.WriteString("  \"suggestions\": [\"<string>\", ...]\n")
	sb.WriteString("}\n")

	return sb.String()
}
