package services

import (
	"context"
	"encoding/json"
	"fmt"
	"main/errs"
	"main/internal/ai"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AIService interface {
	// AnalyzeException(ctx context.Context, orderID int64, notes string) (*dto.AnalyzeExceptionResponse, error)
	AnalyzeException(ctx context.Context, orderID int64) (*dto.AnalyzeExceptionResponse, error)
	GetLatestAnalysis(ctx context.Context, orderID int64) (*dto.AnalyzeExceptionResponse, error)
	GenerateDraft(ctx context.Context, req dto.GenerateDraftAPIRequest) (*dto.GenerateDraftAPIResponse, error)
	TriggerEvaluation(ctx context.Context, req dto.TriggerEvaluationRequest) (*dto.TriggerEvaluationResponse, error)
	GetEvaluationRun(ctx context.Context, runID int64) (*dto.GetEvaluationRunResponse, error)
	GetEvaluationDetails(ctx context.Context, runID int64) (*dto.GetEvaluationDetailsResponse, error)
	ReloadKnowledge(ctx context.Context) error
}

type aiService struct {
	aiRepo         repositories.AIRepository
	analyzer       *ai.ExceptionAnalyzer
	draftGenerator *ai.DraftGenerator
	aiDraftRepo    repositories.AIDraftRepository
	evalRepo       repositories.AIEvaluationRepository
}

func NewAIService(aiRepo repositories.AIRepository, analyzer *ai.ExceptionAnalyzer, draftGenerator *ai.DraftGenerator, aiDraftRepo repositories.AIDraftRepository, evalRepo repositories.AIEvaluationRepository) AIService {
	return &aiService{
		aiRepo:         aiRepo,
		analyzer:       analyzer,
		draftGenerator: draftGenerator,
		aiDraftRepo:    aiDraftRepo,
		evalRepo:       evalRepo,
	}
}

func (s *aiService) ReloadKnowledge(ctx context.Context) error {
	return s.analyzer.ReloadKnowledge(ctx)
}

// func (s *aiService) AnalyzeException(ctx context.Context, orderID int64, notes string) (*dto.AnalyzeExceptionResponse, error) {
func (s *aiService) AnalyzeException(ctx context.Context, orderID int64) (*dto.AnalyzeExceptionResponse, error) {

	aiCtx, err := s.aiRepo.GetAIContextByOrderID(ctx, orderID)
	if err != nil {
		return nil, errs.ERR_NOT_FOUND
	}

	// result, err := s.analyzer.Analyze(ctx, aiCtx, notes)
	result, err := s.analyzer.Analyze(ctx, aiCtx)
	if err != nil {
		return nil, fmt.Errorf("Error in Analyze Service: %w", err)
	}

	exception := mapResultToModel(orderID, result)
	if saveErr := s.aiRepo.Save(ctx, exception); saveErr != nil {
		return nil, fmt.Errorf("Error in save AI result: %w", saveErr)
	}

	return mapToResponse(exception), nil
}

func (s *aiService) GetLatestAnalysis(ctx context.Context, orderID int64) (*dto.AnalyzeExceptionResponse, error) {
	exception, err := s.aiRepo.GetLatestAnalysisByOrderID(ctx, orderID)
	if err != nil {
		return nil, errs.ERR_NOT_FOUND
	}

	resp := mapToResponse(exception)

	// Fetch latest customer update draft for this order (best-effort, not required)
	// draft, draftErr := s.aiDraftRepo.GetLatestByOrderID(ctx, orderID)
	// if draftErr == nil && draft != nil {
	// 	resp.CustomerUpdateDraft = draft.DraftMessage
	// }

	return resp, nil
}

func (s *aiService) GenerateDraft(ctx context.Context, req dto.GenerateDraftAPIRequest) (*dto.GenerateDraftAPIResponse, error) {

	aiCtx, err := s.aiRepo.GetAIContextByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, errs.ERR_NOT_FOUND
	}

	tone := strings.ToLower(strings.TrimSpace(req.Tone))
	switch tone {
	case models.DraftToneApologetic, models.DraftToneInformative, models.DraftToneProactive, models.DraftToneNeutral:
	case "empathetic":
		tone = models.DraftToneApologetic
	default:
		tone = models.DraftToneNeutral
	}

	channel := strings.ToLower(strings.TrimSpace(req.Channel))
	switch channel {
	case "sms", "email", "push":
	default:
		channel = "email"
	}

	lastestException, err := s.aiRepo.GetLatestAnalysisByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, errs.ERR_NOT_FOUND
	}

	adapterInput := dto.CustomerUpdateDraftInput{
		OrderID:         req.OrderID,
		CustomerName:    aiCtx.CustomerName,
		ShippingAddress: aiCtx.ShippingAddress,
		CurrentStatus:   (string)(aiCtx.CurrentStatus),
		LikelyReason:    lastestException.LikelyReason,
		ExceptionType:   lastestException.ExceptionType,
		Tone:            tone,
		Channel:         channel,
	}

	result, err := s.draftGenerator.Generate(ctx, adapterInput)
	if err != nil {
		return nil, errs.ERR_AI_GENERATE_CONTENT_FAILED
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

	confidence := result.ConfidenceScore
	reqID := uuid.NewString()
	draft := &models.AICustomerUpdateDraft{
		OrderID:               req.OrderID,
		AIExceptionResultID:   &lastestException.ID,
		DraftMessage:          result.CustomerUpdateDraft,
		Tone:                  tone,
		Channel:               channel,
		ConfidenceScore:       &confidence,
		FallbackUsed:          result.FallbackUsed,
		FallbackReason:        fallbackReason,
		RawResponse:           rawResponse,
		PromptTemplateVersion: ai.PromptTemplateVersion,
		DurationMs:            durationMs,
		ReviewStatus:          models.DraftReviewStatusPending,
		RequestID:             &reqID,
	}

	if saveErr := s.aiDraftRepo.Save(ctx, draft); saveErr != nil {
		return nil, errs.ERR_INTERNAL_SERVER
	}

	return &dto.GenerateDraftAPIResponse{
		OrderID:               req.OrderID,
		DraftMessage:          result.CustomerUpdateDraft,
		Tone:                  tone,
		ConfidenceScore:       result.ConfidenceScore,
		FallbackUsed:          result.FallbackUsed,
		PromptTemplateVersion: ai.PromptTemplateVersion,
		GeneratedAt:           time.Now(),
	}, nil
}

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
		var js interface{}
		if err := json.Unmarshal([]byte(result.RawResponse), &js); err == nil {
			rawResponse = datatypes.JSON(result.RawResponse)
		} else {
			if bytes, marshalErr := json.Marshal(result.RawResponse); marshalErr == nil {
				rawResponse = datatypes.JSON(bytes)
			}
		}
	}

	reqID := uuid.NewString()

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
		RequestID:             &reqID,
		EvaluatedAt:           now,
	}
}

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
		FallbackReason:        e.FallbackReason,
		PromptTemplateVersion: e.PromptTemplateVersion,
		EvaluatedAt:           e.EvaluatedAt,
	}
}

func (s *aiService) TriggerEvaluation(ctx context.Context, req dto.TriggerEvaluationRequest) (*dto.TriggerEvaluationResponse, error) {
	// 1. Read evaluation_cases.json
	datasetFile := "evaluation_cases.json"
	bytes, err := os.ReadFile(datasetFile)
	if err != nil {
		return nil, fmt.Errorf("could not read dataset file: %w", err)
	}

	var dataset dto.EvaluationDataset
	if err := json.Unmarshal(bytes, &dataset); err != nil {
		return nil, fmt.Errorf("could not parse dataset JSON: %w", err)
	}

	// fmt.Printf("Number of cases: %d\n", len(dataset.Cases))
	// jsonBytes, _ := json.MarshalIndent(dataset, "", "  ")
	// fmt.Printf("Cases: %s\n", string(jsonBytes))

	// 2. Create Run Record in DB (Status = PENDING)
	runRecord := &models.AIEvaluationRun{
		DatasetName: "evaluation_cases.json",
		TotalCases:  len(dataset.Cases),
		Status:      models.EVAL_STATUS_PENDING,
	}
	if err := s.evalRepo.CreateRun(ctx, runRecord); err != nil {
		return nil, fmt.Errorf("could not create evaluation run: %w", err)
	}

	// 3. Initialize Worker
	maxWorkers, _ := strconv.Atoi(os.Getenv("EVAL_MAX_WORKERS"))
	worker := NewEvaluationWorker(s.analyzer, s.evalRepo, maxWorkers)

	// 4. Trigger Worker in background goroutine
	go worker.Run(runRecord.ID, dataset.Cases)

	// 5. Return immediate response
	return &dto.TriggerEvaluationResponse{
		RunID: runRecord.ID,
		// Status:  string(runRecord.Status),
		Message: "Batch evaluation started in background",
	}, nil
}

// GetEvaluationRun trả về summary metrics của 1 evaluation run (dùng để poll status / xem kết quả tổng).
func (s *aiService) GetEvaluationRun(ctx context.Context, runID int64) (*dto.GetEvaluationRunResponse, error) {
	run, err := s.evalRepo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, errs.ERR_NOT_FOUND
	}
	return &dto.GetEvaluationRunResponse{
		RunID:         run.ID,
		DatasetName:   run.DatasetName,
		Status:        run.Status,
		TotalCases:    run.TotalCases,
		PassedCases:   run.PassedCases,
		FailedCases:   run.FailedCases,
		FallbackCount: run.FallbackCount,
		AccuracyRate:  run.AccuracyRate,
		AvgLatencyMs:  run.AvgLatencyMs,
		CreatedAt:     run.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     run.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetEvaluationDetails trả về danh sách chi tiết PASS/FAIL của từng test case trong 1 run.
func (s *aiService) GetEvaluationDetails(ctx context.Context, runID int64) (*dto.GetEvaluationDetailsResponse, error) {
	run, err := s.evalRepo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, errs.ERR_NOT_FOUND
	}

	details, err := s.evalRepo.GetDetailsByRunID(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch evaluation details: %w", err)
	}

	items := make([]dto.EvaluationDetailItem, 0, len(details))
	for _, d := range details {
		items = append(items, dto.EvaluationDetailItem{
			ID:             d.ID,
			Status:         d.Status,
			LatencyMs:      d.LatencyMs,
			ExpectedOutput: string(d.ExpectedOutput),
			ActualOutput:   string(d.ActualOutput),
			ErrorMessage:   d.ErrorMessage,
		})
	}

	return &dto.GetEvaluationDetailsResponse{
		RunID:   runID,
		Status:  run.Status,
		Details: items,
	}, nil
}
