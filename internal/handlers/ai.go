package handlers

import (
	"errors"
	"main/constant"
	"main/errs"
	"main/internal/dto"
	"main/internal/services"
	"main/response"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type AIHandler struct {
	aiService services.AIService
}

func NewAIHandler(aiService services.AIService) *AIHandler {
	return &AIHandler{aiService: aiService}
}

// AnalyzeException godoc
// @Summary Analyze order exception using AI
// @Description Analyze order exceptions and provide actionable insights. Uses rule-based fallback if AI is unavailable.
// @Tags AI
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.ResponseStruct{data=dto.AnalyzeExceptionResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 404 {object} response.ErrorNotFoundResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/ai/orders/{id}/exception-analysis [post]
func (h *AIHandler) AnalyzeException(c fiber.Ctx) error {
	// 1. Parse order ID from path
	orderID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || orderID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	// var req dto.AnalyzeExceptionRequest
	// _ = c.Bind().Body(&req)

	// if err := validator.New().Struct(req); err != nil {
	// 	return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	// }

	// result, err := h.aiService.AnalyzeException(c.Context(), orderID, req.Note)
	result, err := h.aiService.AnalyzeException(c.Context(), orderID)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}

// GetLatestAnalysis godoc
// @Summary Get latest AI exception analysis for an order
// @Description Retrieve the most recent AI-generated exception insights for a specific order, including customer update draft.
// @Tags AI
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.ResponseStruct{data=dto.AnalyzeExceptionResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 404 {object} response.ErrorNotFoundResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/ai/orders/{id}/insights/latest [get]
func (h *AIHandler) GetLatestAnalysis(c fiber.Ctx) error {
	orderID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || orderID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	result, err := h.aiService.GetLatestAnalysis(c.Context(), orderID)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}

// GenerateDraft godoc
// @Summary Generate draft message using AI
// @Description Generate draft message for order exception using AI.
// @Tags AI
// @Accept json
// @Produce json
// @Param request body dto.GenerateDraftAPIRequest true "Generate draft message"
// @Success 200 {object} response.ResponseStruct{data=dto.GenerateDraftAPIResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/ai/orders/customer-update-draft [post]
func (h *AIHandler) GenerateDraft(c fiber.Ctx) error {
	var req dto.GenerateDraftAPIRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	if req.OrderID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	result, err := h.aiService.GenerateDraft(c.Context(), req)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}

// TriggerEvaluation godoc
// @Summary Trigger AI batch evaluation
// @Description Start an asynchronous batch evaluation of the AI rules against synthetic ground truth data.
// @Tags AI Evaluation
// @Accept json
// @Produce json
// @Param request body dto.TriggerEvaluationRequest true "Dataset name"
// @Success 202 {object} response.ResponseStruct{data=dto.TriggerEvaluationResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/ai/evaluations/order-exceptions [post]
func (h *AIHandler) TriggerEvaluation(c fiber.Ctx) error {
	var req dto.TriggerEvaluationRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	if req.DatasetName == "" {
		req.DatasetName = "evaluation_cases.json" // Default
	}

	result, err := h.aiService.TriggerEvaluation(c.Context(), req)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 202, constant.SUCCESS.Message, result)
}

// GetEvaluationRun godoc
// @Summary Get evaluation run summary
// @Description Get the summary metrics of an evaluation run (status, passed/failed counts, accuracy). Poll this until status is COMPLETED.
// @Tags AI Evaluation
// @Produce json
// @Param run_id path int true "Evaluation Run ID"
// @Success 200 {object} response.ResponseStruct{data=dto.GetEvaluationRunResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 404 {object} response.ErrorNotFoundResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/ai/evaluations/{run_id} [get]
func (h *AIHandler) GetEvaluationRun(c fiber.Ctx) error {
	runID, err := strconv.ParseInt(c.Params("run_id"), 10, 64)
	if err != nil || runID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	result, err := h.aiService.GetEvaluationRun(c.Context(), runID)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}

// GetEvaluationDetails godoc
// @Summary Get evaluation case details (PASS/FAIL per case)
// @Description Get the detailed PASS/FAIL result of every individual test case within an evaluation run.
// @Tags AI Evaluation
// @Produce json
// @Param run_id path int true "Evaluation Run ID"
// @Success 200 {object} response.ResponseStruct{data=dto.GetEvaluationDetailsResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 404 {object} response.ErrorNotFoundResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/ai/evaluations/{run_id}/details [get]
func (h *AIHandler) GetEvaluationDetails(c fiber.Ctx) error {
	runID, err := strconv.ParseInt(c.Params("run_id"), 10, 64)
	if err != nil || runID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	result, err := h.aiService.GetEvaluationDetails(c.Context(), runID)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}

// ReloadKnowledge godoc
// @Summary Reload Knowledge Base from DB
// @Description Reload knowledge entries from database to memory cache.
// @Tags AI Admin
// @Produce json
// @Success 200 {object} response.ResponseStruct
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/ai/knowledge/reload [post]
func (h *AIHandler) ReloadKnowledge(c fiber.Ctx) error {
	if err := h.aiService.ReloadKnowledge(c.Context()); err != nil {
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}
	return response.ResponseSuccess(c, 200, "Knowledge Base reloaded successfully", nil)
}
