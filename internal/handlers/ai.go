package handlers

import (
	"main/internal/services"

	"github.com/gofiber/fiber/v3"
)

type AIHandler struct {
	aiService services.AIService
}

func NewAIHandler(aiService services.AIService) *AIHandler {
	return &AIHandler{aiService: aiService}
}

func (h *AIHandler) AnalyzeException(c fiber.Ctx) error {
	return nil
}
