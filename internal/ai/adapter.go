package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"main/errs"
	"main/internal/dto"
	"main/pkg/aiclient"
)

type AIAdapter interface {
	AnalyzeException(ctx context.Context, input dto.ExceptionInput) (string, error)
	SummarizeReport(ctx context.Context, input dto.ExceptionOutput) (dto.ReportSummaryOutput, error)
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
) (string, error) {
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
		OrderID:       input.OrderID,
		CurrentStatus: input.CurrentStatus,
		DriverNotes:   input.ErrorMessage,
		EventTimeline: timeline,
		AnalyzedAt:    time.Now().Format(time.RFC3339),
	}
	SanitizePromptContext(&promptCtx)

	prompt := BuildExceptionAnalysisPrompt(promptCtx)
	// fmt.Printf("[DEBUG][adapter.AnalyzeException] Built prompt (len=%d)\n", len(prompt))

	rawText, err := g.client.GenerateContent(ctx, prompt)
	if err != nil {
		fmt.Printf("[DEBUG][adapter.AnalyzeException] GenerateContent error: %v\n", err)
		return "", errs.ERR_GEMINI_GENERATE_CONTENT_FAILED
	}

	// fmt.Printf("[DEBUG][adapter.AnalyzeException] Raw AI response (len=%d): %s\n", len(rawText), rawText)

	// validated, err := ParseAndValidateAIOutput(rawText)
	// if err != nil {
	// 	fmt.Printf("[DEBUG][adapter.AnalyzeException] ParseAndValidateAIOutput error: %v\n", err)
	// 	return "", errs.ERR_AI_RESPONSE_VALIDATION_FAILED
	// }

	// return dto.ExceptionOutput{
	// 	ExceptionType:      "",
	// 	Severity:           "",
	// 	LikelyReason:       "",
	// 	InternalNextAction: "",
	// 	Suggestion:         "",
	// 	Confidence:         0,
	// }, nil

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
	fmt.Println("[DEBUG][adapter.Ping] Sending ping to Gemini...")
	_, err := g.client.GenerateContent(ctx, "ping")
	if err != nil {
		fmt.Printf("[DEBUG][adapter.Ping] Ping FAILED: %v\n", err)
		return errs.ERR_AI_PING_FAILED
	}
	fmt.Println("[DEBUG][adapter.Ping] Ping OK")
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
