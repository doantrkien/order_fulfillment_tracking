package handlers

import (
	"errors"
	"main/constant"
	"main/errs"
	"main/internal/dto"
	"main/internal/services"
	"main/response"
	"strconv"

	"github.com/go-playground/validator"
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
// @Param request body dto.AnalyzeExceptionRequest false "Optional analysis context notes"
// @Success 200 {object} response.ResponseStruct{data=dto.AnalyzeExceptionResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 404 {object} response.ErrorNotFoundResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/ai/orders/{id}/exception-analysis [post]
func (h *AIHandler) AnalyzeException(c fiber.Ctx) error {
	orderID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || orderID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	var req dto.AnalyzeExceptionRequest
	_ = c.Bind().Body(&req)

	if err := validator.New().Struct(req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	requestID, _ := c.Locals("requestid").(string)

	result, err := h.aiService.AnalyzeException(c.Context(), orderID, req.Note, requestID)
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

	requestID, _ := c.Locals("requestid").(string)

	result, err := h.aiService.GenerateDraft(c.Context(), req, requestID)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}

// RunEvaluation godoc
// @Summary Run evaluation on order exceptions
// @Description Run evaluation on order exceptions using provided cases.
// @Tags AI
// @Accept json
// @Produce json
// @Param request body dto.EvaluationRequest true "Evaluation input"
// @Success 200 {object} response.ResponseStruct{data=dto.EvaluationResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/ai/evaluations/order-exceptions [post]
func (h *AIHandler) RunEvaluation(c fiber.Ctx) error {
	var req dto.EvaluationRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}
	if len(req.Cases) < 20 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	result, err := h.aiService.RunEvaluation(c.Context(), req)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}
	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}
