package handlers

import (
	"errors"
	"main/internal/dto"
	"main/internal/services"
	"main/pkg/utils/constant"
	"main/pkg/utils/errs"
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
	var query dto.OrderQuery

	if err := c.Bind().Query(&query); err != nil {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}

	result, total, err := h.orderService.GetAllOrder(query)
	if err != nil {
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	return response.PaginatedSuccess(
		c,
		constant.SUCCESS,
		result,
		query.Page,
		query.Limit,
		total,
	)
}

func (h *OrderHandler) GetOrderDetail(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	order, err := h.orderService.GetOrder(id)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.Reponse(c, 404, constant.NOT_FOUND, nil)
		}
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	return response.Reponse(c, 200, constant.SUCCESS, order)
}

func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
	var req dto.OrderRequest

	if err := c.Bind().Body(&req); err != nil {
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
	var req dto.UpdateStatusRequest

	if err != nil {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	if err := c.Bind().Body(&req); err != nil {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	if req.Status == "" {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	_, err = h.orderService.UpdateOrderStatus(id, string(req.Status))
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.Reponse(c, 404, constant.NOT_FOUND, nil)
		}

		if errors.Is(err, errs.ORDER_STATUS_TRANSITION_INVALID) {
			return response.Reponse(c, 400, constant.INVALID_STATUS, nil)
		}

		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	return response.Reponse(c, 200, constant.SUCCESS, nil)
}
