package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"main/errs"
	"main/internal/ai"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"

	"gorm.io/datatypes"
)

// AIService defines the business-level interface for AI features.
type AIService interface {
	AnalyzeException(ctx context.Context, orderID int64, notes string) (*dto.AnalyzeExceptionResponse, error)
	GetLatestAnalysis(ctx context.Context, orderID int64, notes string) (*dto.AnalyzeExceptionResponse, error)
	UpdateDraft(ctx context.Context, req dto.UpdateDraftAPIRequest) (*dto.UpdateDraftAPIResponse, error)
}

type aiService struct {
	aiRepo         repositories.AIRepository
	analyzer       *ai.ExceptionAnalyzer
	draftGenerator *ai.DraftGenerator
	aiDraftRepo    repositories.AIDraftRepository
}

// NewAIService creates the AI service wired to the repository and analyzer.
func NewAIService(aiRepo repositories.AIRepository, analyzer *ai.ExceptionAnalyzer, draftGenerator *ai.DraftGenerator, aiDraftRepo repositories.AIDraftRepository) AIService {
	return &aiService{
		aiRepo:         aiRepo,
		analyzer:       analyzer,
		draftGenerator: draftGenerator,
		aiDraftRepo:    aiDraftRepo,
	}
}

func (s *aiService) AnalyzeException(ctx context.Context, orderID int64, notes string) (*dto.AnalyzeExceptionResponse, error) {
	// 1. Fetch order context from DB
	aiCtx, err := s.aiRepo.GetAIContextByOrderID(ctx, orderID)
	if err != nil {
		return nil, errs.ERR_NOT_FOUND
	}

	// 2. Run the analyzer (AI-first with automatic fallback)
	result, err := s.analyzer.Analyze(ctx, aiCtx, notes)
	if err != nil {
		return nil, fmt.Errorf("analyzer error: %w", err)
	}

	// 3. Persist the result to the database
	exception := mapResultToModel(orderID, result)
	if saveErr := s.aiRepo.Save(ctx, exception); saveErr != nil {
		return nil, fmt.Errorf("failed to save AI result: %w", saveErr)
	}

	// 4. Map to response DTO
	return mapToResponse(exception), nil
}

func (s *aiService) GetLatestAnalysis(ctx context.Context, orderID int64, notes string) (*dto.AnalyzeExceptionResponse, error) {
	exception, err := s.aiRepo.GetLatestAnalysisByOrderID(ctx, orderID)
	if err != nil {
		return nil, errs.ERR_NOT_FOUND
	}

	return mapToResponse(exception), nil
}

func (s *aiService) UpdateDraft(ctx context.Context, req dto.UpdateDraftAPIRequest) (*dto.UpdateDraftAPIResponse, error) {
	// 1. Fetch order context to enrich the AI input with customer info
	aiCtx, err := s.aiRepo.GetAIContextByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, errs.ERR_NOT_FOUND
	}

	// Normalize and validate Tone to avoid DB check constraint violations
	tone := strings.ToLower(strings.TrimSpace(req.Tone))
	switch tone {
	case models.DraftToneApologetic, models.DraftToneInformative, models.DraftToneProactive, models.DraftToneNeutral:
		// valid tone
	case "empathetic":
		tone = models.DraftToneApologetic // Map common alternative to allowed value
	default:
		tone = models.DraftToneNeutral
	}

	lastestException, err := s.aiRepo.GetLatestAnalysisByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, errs.ERR_NOT_FOUND
	}

	// 2. Build adapter input
	adapterInput := dto.CustomerUpdateDraftInput{
		OrderID:         req.OrderID,
		CustomerName:    aiCtx.CustomerName,
		ShippingAddress: aiCtx.ShippingAddress,
		CurrentStatus:   (string)(aiCtx.CurrentStatus),
		LikelyReason:    lastestException.LikelyReason,
		ExceptionType:   lastestException.ExceptionType,
		Tone:            tone,
		Channel:         req.Channel,
	}

	// 3. Generate draft — DraftGenerator handles AI call, JSON parsing, and fallback internally
	result, err := s.draftGenerator.Generate(ctx, adapterInput)
	if err != nil {
		return nil, errs.ERR_GEMINI_GENERATE_CONTENT_FAILED
	}

	var fallbackReason *string
	if result.FallbackUsed && result.FallbackReason != "" {
		reason := result.FallbackReason
		fallbackReason = &reason
	}

	var durationMs *int
	if result.DurationMs > 0 {
		d := result.DurationMs
		durationMs = &d
	}

	var rawResponse datatypes.JSON
	if result.RawResponse != "" {
		var js interface{}
		if err := json.Unmarshal([]byte(result.RawResponse), &js); err == nil {
			rawResponse = datatypes.JSON(result.RawResponse)
		} else {
			if bytes, marshalErr := json.Marshal(result.RawResponse); marshalErr == nil {
				rawResponse = datatypes.JSON(bytes)
			}
		}
	}

	// 4. Persist draft to DB
	confidence := result.ConfidenceScore
	draft := &models.AICustomerUpdateDraft{
		OrderID:               req.OrderID,
		AIExceptionResultID:   &lastestException.ID,
		DraftMessage:          result.CustomerUpdateDraft,
		Tone:                  tone,
		ConfidenceScore:       &confidence,
		FallbackUsed:          result.FallbackUsed,
		FallbackReason:        fallbackReason,
		RawResponse:           rawResponse,
		PromptTemplateVersion: ai.PromptTemplateVersion,
		DurationMs:            durationMs,
		ReviewStatus:          models.DraftReviewStatusPending,
	}
	if saveErr := s.aiDraftRepo.Save(ctx, draft); saveErr != nil {
		return nil, errs.ERR_INTERNAL_SERVER
	}

	// 5. Build and return response DTO
	return &dto.UpdateDraftAPIResponse{
		OrderID:               req.OrderID,
		DraftMessage:          result.CustomerUpdateDraft,
		Tone:                  tone,
		ConfidenceScore:       result.ConfidenceScore,
		FallbackUsed:          result.FallbackUsed,
		PromptTemplateVersion: ai.PromptTemplateVersion,
		GeneratedAt:           time.Now(),
	}, nil
}

// stripMarkdownFences removes ```json ... ``` wrappers that some LLMs add.
func stripMarkdownFences(s string) string {
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

// mapResultToModel converts the analyzer output to the database model.
func mapResultToModel(orderID int64, result *ai.AnalysisResult) *models.AIException {
	now := time.Now()

	var fallbackReason *string
	if result.FallbackUsed && result.FallbackReason != "" {
		reason := result.FallbackReason
		fallbackReason = &reason
	}

	var durationMs *int
	if result.DurationMs > 0 {
		d := result.DurationMs
		durationMs = &d
	}

	var rawResponse datatypes.JSON
	if result.RawResponse != "" {
		// Ensure rawResponse is valid JSON before saving to DB JSONB column
		var js interface{}
		if err := json.Unmarshal([]byte(result.RawResponse), &js); err == nil {
			rawResponse = datatypes.JSON(result.RawResponse)
		} else {
			// If not valid JSON, serialize the raw string into a JSON string format
			if bytes, marshalErr := json.Marshal(result.RawResponse); marshalErr == nil {
				rawResponse = datatypes.JSON(bytes)
			}
		}
	}

	return &models.AIException{
		OrderID:               orderID,
		ExceptionType:         result.ExceptionType,
		Severity:              result.Severity,
		LikelyReason:          result.LikelyReason,
		InternalNextAction:    result.InternalNextAction,
		ConfidenceScore:       result.ConfidenceScore,
		FallbackUsed:          result.FallbackUsed,
		FallbackReason:        fallbackReason,
		PromptTemplateVersion: ai.PromptTemplateVersion,
		DurationMs:            durationMs,
		RawResponse:           rawResponse,
		EvaluatedAt:           now,
	}
}

// mapToResponse converts the database model to the API response DTO.
func mapToResponse(e *models.AIException) *dto.AnalyzeExceptionResponse {
	fallbackReason := ""
	if e.FallbackReason != nil {
		fallbackReason = *e.FallbackReason
	}
	return &dto.AnalyzeExceptionResponse{
		ResultID:              fmt.Sprintf("res-%d", e.ID),
		OrderID:               fmt.Sprintf("%d", e.OrderID),
		ExceptionType:         e.ExceptionType,
		Severity:              e.Severity,
		LikelyReason:          e.LikelyReason,
		InternalNextAction:    e.InternalNextAction,
		ConfidenceScore:       e.ConfidenceScore,
		FallbackUsed:          e.FallbackUsed,
		FallbackReason:        fallbackReason,
		PromptTemplateVersion: e.PromptTemplateVersion,
		EvaluatedAt:           e.EvaluatedAt,
	}
}
