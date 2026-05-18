package handlers

import (
	"main/internal/dto"
	"main/internal/services"
	"main/pkg/utils/response"

	"github.com/gofiber/fiber/v3"
)

type OrderEventHandler struct {
	orderEventService services.OrderEventService
}

func NewOrderEventHandler(orderEventService services.OrderEventService) *OrderEventHandler {
	return &OrderEventHandler{orderEventService: orderEventService}
}

func (h *OrderEventHandler) ImportOrderEvents(c fiber.Ctx) error {
	var req []dto.ImportOrderEventRequest

	if err := c.Bind().Body(&req); err != nil {
		return response.Reponse(c, 400, "invalid payload", nil)
	}

	result, err := h.orderEventService.ImportOrderEvents(req)
	if err != nil {
		return response.Reponse(c, 500, err.Error(), result)
	}

	return response.Reponse(c, 200, "batch processed", result)
}
