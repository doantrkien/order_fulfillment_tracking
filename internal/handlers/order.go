package handlers

import (
	"main/internal/services"
	"main/pkg/utils/constant"
	"main/pkg/utils/response"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type OrderHandler struct {
	orderService services.OrderService
}

func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) GetAllOrder(c fiber.Ctx) error {
	orders, err := h.orderService.GetAllOrder()
	if err != nil {
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	return response.Reponse(c, 200, constant.SUCCESS, orders)
}

func (h *OrderHandler) GetOrderDetail(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	order, err := h.orderService.GetOrder(id)
	if err != nil {
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	return response.Reponse(c, 200, constant.SUCCESS, order)
}
