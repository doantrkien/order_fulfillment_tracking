package handlers

import (
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

func (h *AIHandler) AnalyzeException(c fiber.Ctx) error {
	_, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	var req dto.AnalyzeExceptionRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	return nil
}
