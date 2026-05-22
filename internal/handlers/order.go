package handlers

import (
	"errors"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/services"
	"main/pkg/metrics"
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
		metrics.OrderCreatedTotal.WithLabelValues("error").Inc()
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	_, err := h.orderService.CreateOrder(req)
	if err != nil {
		metrics.OrderCreatedTotal.WithLabelValues("error").Inc()
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	metrics.OrderCreatedTotal.WithLabelValues("success").Inc()
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

	order, err := h.orderService.UpdateOrderStatus(id, status)
	if err != nil {
		if errors.Is(err, constant.ERR_NOT_FOUND) {
			metrics.OrderStatusUpdatedTotal.WithLabelValues(status, "not_found").Inc()
			return response.Reponse(c, 404, constant.NOT_FOUND, nil)
		}
		metrics.OrderStatusUpdatedTotal.WithLabelValues(status, "error").Inc()
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	if models.IsValidTransition(order.CurrentStatus, models.OrderStatus(status)) {
		metrics.OrderStatusUpdatedTotal.WithLabelValues(status, "invalid_transition").Inc()
		return response.Reponse(c, 400, constant.INVALID_STATUS, nil)
	}

	metrics.OrderStatusUpdatedTotal.WithLabelValues(status, "success").Inc()
	return response.Reponse(c, 200, constant.SUCCESS, nil)
}
