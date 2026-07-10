package ai

import (
	"context"
	"encoding/json"
	"time"

	"main/errs"
	dto_ai "main/internal/dto/ai"

	"main/pkg/aiclient"
	"main/utils/helpers"
)

type AIAdapter interface {
	AnalyzeException(ctx context.Context, input dto_ai.ExceptionPromptContext) (dto_ai.AIAnalysisResult, string, error)
	DraftCustomerUpdate(ctx context.Context, input dto_ai.CustomerUpdateDraftInput) (string, error)
	ReloadKnowledge(ctx context.Context) error
}

type aiAdapter struct {
	client  aiclient.AIClient
	kbStore *KnowledgeStore
}

func NewAIAdapter(client aiclient.AIClient, kbStore *KnowledgeStore) *aiAdapter {
	return &aiAdapter{client: client, kbStore: kbStore}
}

func (g *aiAdapter) AnalyzeException(
	ctx context.Context,
	input dto_ai.ExceptionPromptContext,
) (dto_ai.AIAnalysisResult, string, error) {
	timeline := make([]dto_ai.EventTimelineEntry, 0, len(input.EventTimeline))
	for _, e := range input.EventTimeline {
		timeline = append(timeline, dto_ai.EventTimelineEntry{
			FromStatus: e.FromStatus,
			ToStatus:   e.ToStatus,
			UpdatedBy:  e.UpdatedBy,
			EventAt:    e.EventAt,
		})
	}

	promptCtx := dto_ai.ExceptionPromptContext{
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

	var knowledge []KnowledgeEntry
	if g.kbStore != nil {
		knowledge = g.kbStore.ClassifyDriverNote(ctx, input.DriverNotes)
	} else {
		knowledge = GetKnowledgeBase(input.DriverNotes)
	}
	prompt := BuildExceptionAnalysisPrompt(promptCtx, knowledge)

	rawText, err := g.client.GenerateContent(ctx, prompt)
	// fmt.Println("Raw Text Ai", rawText)
	if err != nil {
		return dto_ai.AIAnalysisResult{}, "", errs.ERR_AI_GENERATE_CONTENT_FAILED
	}

	var output dto_ai.AIAnalysisResult
	cleaned := helpers.StripFences(rawText)
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return dto_ai.AIAnalysisResult{}, rawText, errs.ERR_AI_RESPONSE_VALIDATION_FAILED
	}

	return output, rawText, nil
}

func (g *aiAdapter) DraftCustomerUpdate(
	ctx context.Context,
	input dto_ai.CustomerUpdateDraftInput,
) (string, error) {
	SanitizeCustomerUpdateDraftInput(&input)
	prompt := BuildCustomerUpdateDraftPrompt(input)

	rawText, err := g.client.GenerateContent(ctx, prompt)
	if err != nil {
		return "", errs.ERR_AI_GENERATE_CONTENT_FAILED
	}

	return rawText, nil
}

func (g *aiAdapter) ReloadKnowledge(ctx context.Context) error {
	if g.kbStore != nil {
		return g.kbStore.Reload(ctx)
	}
	return nil
}
