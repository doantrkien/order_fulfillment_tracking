package handlers

import (
	"errors"
	"main/internal/dto"
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

func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
	var req dto.OrderRequest

	if err := c.Bind().Body(&req); err != nil {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	if req.CustomerID <= 0 || req.TotalAmount <= 0 {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	_, err := h.orderService.CreateOrder(req)
	if err != nil {
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	return response.Reponse(c, 201, constant.SUCCESS, nil)
}

func (h *OrderHandler) UpdateOrderStatus(c fiber.Ctx) error {

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	status := c.Query("status")
	if status == "" {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	_, err = h.orderService.UpdateOrderStatus(id, status)
	if err != nil {
		if errors.Is(err, constant.ERR_NOT_FOUND) {
			return response.Reponse(c, 404, constant.NOT_FOUND, nil)
		}
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	return response.Reponse(c, 200, constant.SUCCESS, nil)
}
