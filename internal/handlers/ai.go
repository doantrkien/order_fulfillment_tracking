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

// AnalyzeException handles POST /ai/orders/:id/exception-analysis.
// This endpoint never returns 500 for AI failures — it gracefully
// falls back to rule-based detection with fallback_used: true.
func (h *AIHandler) AnalyzeException(c fiber.Ctx) error {
	// 1. Parse order ID from path
	orderID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || orderID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	// 2. Parse optional request body (notes)
	var req dto.AnalyzeExceptionRequest
	// Body is optional, ignore parse errors for empty body
	_ = c.Bind().Body(&req)

	// 3. Call service (AI-first with automatic fallback)
	result, err := h.aiService.AnalyzeException(c.Context(), orderID, req.Note)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	// 4. Return success — never 500 for AI failures
	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}
