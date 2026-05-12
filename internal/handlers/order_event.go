package handlers

import (
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
	// TODO
	return c.Status(200).JSON(fiber.Map{
		"status":  200,
		"message": "success",
		"data":    nil,
	})
}
