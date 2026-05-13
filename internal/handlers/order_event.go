package handlers

import (
	"main/internal/models"
	"main/internal/services"

	"github.com/gofiber/fiber/v3"
)

type OrderEventHandler struct {
	orderEventService services.OrderEventService
}

func NewOrderEventHandler(orderEventService services.OrderEventService) *OrderEventHandler {
	return &OrderEventHandler{
		orderEventService: orderEventService,
	}
}

func (h *OrderEventHandler) ImportOrderEvents(c fiber.Ctx) error {
	var req []models.OrderEvent

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid payload",
		})
	}

	result, err := h.orderEventService.ImportOrderEvents(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":     err.Error(),
			"accepted":  result.Accepted,
			"rejected":  result.Rejected,
			"duplicate": result.Duplicate,
		})
	}

	return c.JSON(fiber.Map{
		"message":   "batch processed",
		"accepted":  result.Accepted,
		"rejected":  result.Rejected,
		"duplicate": result.Duplicate,
	})
}
