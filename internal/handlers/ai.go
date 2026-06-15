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

func (h *AIHandler) UpdateDraf(c fiber.Ctx) error {
	var req dto.UpdateDraftAPIRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	if req.OrderID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	// result, err := h.aiService.UpdateDraft(c.Context(), req)
	// if err != nil {
	// 	if errors.Is(err, errs.ERR_NOT_FOUND) {
	// 		return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
	// 	}
	// 	return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	// }

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}
