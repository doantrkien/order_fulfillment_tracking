package ai

import (
	"context"
	"encoding/json"
	"time"

	"main/errs"
	"main/internal/dto"
	"main/pkg/aiclient"
)

type AIAdapter interface {
	AnalyzeException(ctx context.Context, input dto.ExceptionInput) (dto.ExceptionOutput, string, error)
	DraftCustomerUpdate(ctx context.Context, input dto.CustomerUpdateDraftInput) (string, error)
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
		DriverNotes:     input.DriverNotes,
		EventTimeline:   timeline,
		AnalyzedAt:      time.Now().Format(time.RFC3339),
	}
	// fmt.Println("Prmot", promptCtx)

	SanitizePromptContext(&promptCtx)

	knowledge := ClassifyDriverNote(input.DriverNotes)
	prompt := BuildExceptionAnalysisPrompt(promptCtx, knowledge)

	rawText, err := g.client.GenerateContent(ctx, prompt)
	// fmt.Println("Raw Text Ai", rawText)
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
