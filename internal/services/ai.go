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
	"strings"
	"time"

	"gorm.io/datatypes"
)

type AIService interface {
	AnalyzeException(ctx context.Context, orderID int64, notes string) (*dto.AnalyzeExceptionResponse, error)
	GetLatestAnalysis(ctx context.Context, orderID int64) (*dto.AnalyzeExceptionResponse, error)
	GenerateDraft(ctx context.Context, req dto.GenerateDraftAPIRequest) (*dto.GenerateDraftAPIResponse, error)
	RunEvaluation(ctx context.Context, req dto.EvaluationRequest) (*dto.EvaluationResponse, error)
}

type aiService struct {
	aiRepo         repositories.AIRepository
	analyzer       *ai.ExceptionAnalyzer
	draftGenerator *ai.DraftGenerator
	aiDraftRepo    repositories.AIDraftRepository
	aiEvalRepo     repositories.AIEvaluationRepository
}

func NewAIService(aiRepo repositories.AIRepository, analyzer *ai.ExceptionAnalyzer, draftGenerator *ai.DraftGenerator, aiDraftRepo repositories.AIDraftRepository, aiEvalRepo repositories.AIEvaluationRepository) AIService {
	return &aiService{
		aiRepo:         aiRepo,
		analyzer:       analyzer,
		draftGenerator: draftGenerator,
		aiDraftRepo:    aiDraftRepo,
		aiEvalRepo:     aiEvalRepo,
	}
}

func (s *aiService) AnalyzeException(ctx context.Context, orderID int64, notes string) (*dto.AnalyzeExceptionResponse, error) {

	aiCtx, err := s.aiRepo.GetAIContextByOrderID(ctx, orderID)
	if err != nil {
		return nil, errs.ERR_NOT_FOUND
	}

	result, err := s.analyzer.Analyze(ctx, aiCtx, notes)
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
		Channel:         req.Channel,
	}

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

func (s *aiService) RunEvaluation(ctx context.Context, req dto.EvaluationRequest) (*dto.EvaluationResponse, error) {
	var (
		results       []dto.EvaluationCaseResult
		passedCases   int
		fallbackCases int
		totalConf     float64
	)

	for _, c := range req.Cases {
		aiCtx, err := buildAIContextForEval(ctx, s.aiRepo, &c)
		if err != nil {
			results = append(results, dto.EvaluationCaseResult{
				CaseID:     c.CaseID,
				Passed:     false,
				FailReason: fmt.Sprintf("context error: %s", err.Error()),
			})
			continue
		}

		// 2. Gọi AI analyzer
		result, err := s.analyzer.Analyze(ctx, aiCtx, "")
		if err != nil {
			results = append(results, dto.EvaluationCaseResult{
				CaseID:     c.CaseID,
				Passed:     false,
				FailReason: fmt.Sprintf("analyzer error: %s", err.Error()),
			})
			continue
		}

		// 3. So sánh kết quả với expected
		passed := strings.EqualFold(result.Severity, c.ExpectedSeverity) &&
			strings.EqualFold(result.ExceptionType, c.ExpectedExceptionType)

		if passed {
			passedCases++
		}
		if result.FallbackUsed {
			fallbackCases++
		}
		totalConf += result.ConfidenceScore

		results = append(results, dto.EvaluationCaseResult{
			CaseID:              c.CaseID,
			Passed:              passed,
			FallbackUsed:        result.FallbackUsed,
			ActualSeverity:      result.Severity,
			ActualExceptionType: result.ExceptionType,
			ConfidenceScore:     result.ConfidenceScore,
		})
	}

	// 4. Tính aggregate
	total := len(req.Cases)
	passRate := 0.0
	avgConf := 0.0
	if total > 0 {
		passRate = float64(passedCases) / float64(total)
		avgConf = totalConf / float64(total)
	}

	// 5. Lưu vào ai_evaluation_runs — dùng s.aiEvalRepo (không phải s.evalRepo)
	env := req.Environment
	if env == "" {
		env = models.EvalEnvironmentDev
	}

	rawJSON, _ := json.Marshal(results)
	runBy := req.RunBy

	run := &models.AIEvaluationRun{
		WorkflowName:          models.WorkflowNameOrderException,
		PromptTemplateVersion: ai.PromptTemplateVersion,
		RunBy:                 &runBy,
		Environment:           env,
		TriggeredBy:           models.EvalTriggeredByManual,
		TotalCases:            total,
		PassedCases:           passedCases,
		FailedCases:           total - passedCases,
		FallbackCases:         fallbackCases,
		AvgConfidence:         &avgConf,
		PassRate:              &passRate,
		RawResults:            datatypes.JSON(rawJSON),
		RunAt:                 time.Now(),
	}

	if err := s.aiEvalRepo.Save(ctx, run); err != nil { // <-- aiEvalRepo bukan evalRepo
		return nil, fmt.Errorf("failed to save evaluation run: %w", err)
	}

	// 6. Return response — chỉ dùng field có trong dto.EvaluationResponse hiện tại
	return &dto.EvaluationResponse{
		WorkflowName:  models.WorkflowNameOrderException,
		TotalCases:    total,
		PassedCases:   passedCases,
		FailedCases:   total - passedCases,
		FallbackCases: fallbackCases,
		PassRate:      passRate,
		AvgConfidence: avgConf,
		Results:       results,
		RunAt:         run.RunAt,
	}, nil
}

func buildAIContextForEval(ctx context.Context, aiRepo repositories.AIRepository, c *dto.EvaluationCase) (*models.AIContext, error) {
	if c.OrderID > 0 {
		return aiRepo.GetAIContextByOrderID(ctx, c.OrderID)
	}
	if c.SyntheticInput != nil {
		return buildAIContextFromSynthetic(c.SyntheticInput), nil
	}
	return nil, fmt.Errorf("case %s: must provide either order_id or synthetic_input", c.CaseID)
}

func buildAIContextFromSynthetic(input *dto.ExceptionInput) *models.AIContext {
	events := make([]models.AIEvent, 0, len(input.EventHistory))
	for _, e := range input.EventHistory {
		parsedAt, _ := time.Parse(time.RFC3339, e.EventAt)
		events = append(events, models.AIEvent{
			EventAt:        parsedAt,
			PreviousStatus: models.OrderStatus(e.FromStatus),
			NewStatus:      models.OrderStatus(e.ToStatus),
			UpdatedBy:      e.UpdatedBy,
		})
	}

	// Parse CreatedAt from synthetic input; fall back to zero time if missing
	createdAt, _ := time.Parse(time.RFC3339, input.CreatedAt)

	// Map ErrorMessage to DriverNotes so rule-based detection can inspect it
	driverNotes := input.ErrorMessage

	status := models.OrderStatus(input.CurrentStatus)
	return &models.AIContext{
		OrderID:         input.OrderID,
		CreatedAt:       createdAt,
		CurrentStatus:   status,
		TotalAmount:     input.TotalAmount,
		CustomerName:    input.CustomerName,
		ShippingAddress: input.ShippingAddress,
		PaymentStatus:   models.DerivePaymentStatus(status),
		RefundStatus:    models.DeriveRefundStatus(status),
		DriverNotes:     driverNotes,
		Events:          events,
	}
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
