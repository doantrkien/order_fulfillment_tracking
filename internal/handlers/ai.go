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

	result, err := h.aiService.AnalyzeException(c.Context(), orderID, req.Note)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}

func (h *AIHandler) GetLatestAnalysis(c fiber.Ctx) error {
	orderID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || orderID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	var req dto.AnalyzeExceptionRequest
	_ = c.Bind().Body(&req)

	result, err := h.aiService.GetLatestAnalysis(c.Context(), orderID, req.Note)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}

// UpdateDraft godoc
// @Summary Generate draft message using AI
// @Description Generate draft message for order exception using AI.
// @Tags AI
// @Accept json
// @Produce json
// @Param request body dto.UpdateDraftAPIRequest true "Generate draft message"
// @Success 200 {object} response.ResponseStruct{data=dto.UpdateDraftAPIResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/ai/orders/customer-update-draft [post]
func (h *AIHandler) UpdateDraf(c fiber.Ctx) error {
	var req dto.UpdateDraftAPIRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	if req.OrderID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	result, err := h.aiService.UpdateDraft(c.Context(), req)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}
