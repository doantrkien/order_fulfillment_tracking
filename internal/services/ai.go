package services

import (
	"context"
	"fmt"
	"main/errs"
	"main/internal/ai"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"
	"time"

	"gorm.io/datatypes"
)

// AIService defines the business-level interface for AI exception analysis.
type AIService interface {
	AnalyzeException(ctx context.Context, orderID int64, notes string) (*dto.AnalyzeExceptionResponse, error)
}

type aiService struct {
	aiRepo   repositories.AIRepository
	analyzer *ai.ExceptionAnalyzer
}

// NewAIService creates the AI service wired to the repository and analyzer.
func NewAIService(aiRepo repositories.AIRepository, analyzer *ai.ExceptionAnalyzer) AIService {
	return &aiService{
		aiRepo:   aiRepo,
		analyzer: analyzer,
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
	fmt.Printf("[Debug Service] Analyzed exception result: %+v\n", result)

	// 3. Persist the result to the database
	exception := mapResultToModel(orderID, result)
	if saveErr := s.aiRepo.Save(ctx, exception); saveErr != nil {
		return nil, fmt.Errorf("failed to save AI result: %w", saveErr)
	}

	// fmt.Printf("[Debug Service] Analyzed exception for order %+v\n", exception)

	// 4. Map to response DTO
	return mapToResponse(exception), nil
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
		rawResponse = datatypes.JSON(result.RawResponse)
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
	return &dto.AnalyzeExceptionResponse{
		ResultID:              fmt.Sprintf("res-%d", e.ID),
		OrderID:               fmt.Sprintf("%d", e.OrderID),
		ExceptionType:         e.ExceptionType,
		Severity:              e.Severity,
		LikelyReason:          e.LikelyReason,
		InternalNextAction:    e.InternalNextAction,
		ConfidenceScore:       e.ConfidenceScore,
		FallbackUsed:          e.FallbackUsed,
		PromptTemplateVersion: e.PromptTemplateVersion,
		EvaluatedAt:           e.EvaluatedAt,
	}
}
