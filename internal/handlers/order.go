package handlers

import (
	"main/internal/services"

	"github.com/gofiber/fiber/v3"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) GetAllOrder(c *fiber.Ctx) {
	return
}

func (h *OrderHandler) GetOrderDetail(c *fiber.Ctx) {
	return
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) {

	return
}

func (h *OrderHandler) UpdateOrder(c *fiber.Ctx) {

	return
}
