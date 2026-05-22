package handlers

import (
	"errors"
	"main/internal/dto"
	"main/internal/services"
	"main/pkg/metrics"
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

// GetAllOrder godoc
// @Summary Get all orders
// @Tags Order
// @Accept json
// @Produce json
// @Param status query string false "Order Status"
// @Param customer_name query string false "Customer Name"
// @Param ordered_at query string false "Ordered Date"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} response.PaginatedResponse{data=[]dto.OrderReponse}
// @Failure 400 {object} response.ResponseStruct
// @Failure 401 {object} response.ResponseStruct
// @Failure 403 {object} response.ResponseStruct
// @Failure 500 {object} response.ResponseStruct
// @Security ApiKeyAuth
// @Router /api/v1/orders [get]
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

// GetOrderDetail godoc
// @Summary Get order detail
// @Description Retrieve detail of an order
// @Tags Order
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.ResponseStruct{data=dto.OrderReponse}
// @Failure 400 {object} response.ResponseStruct
// @Failure 401 {object} response.ResponseStruct
// @Failure 403 {object} response.ResponseStruct
// @Failure 404 {object} response.ResponseStruct
// @Failure 500 {object} response.ResponseStruct
// @Security ApiKeyAuth
// @Router /api/v1/orders/{id} [get]
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

// CreateOrder godoc
// @Summary Create order
// @Description Create a new order
// @Tags Order
// @Accept json
// @Produce json
// @Param request body dto.OrderRequest true "Order details"
// @Success 201 {object} response.ResponseStruct
// @Failure 400 {object} response.ResponseStruct
// @Failure 401 {object} response.ResponseStruct
// @Failure 403 {object} response.ResponseStruct
// @Failure 500 {object} response.ResponseStruct
// @Security ApiKeyAuth
// @Router /api/v1/orders [post]
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

// UpdateOrderStatus godoc
// @Summary Update order status
// @Tags Order
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param status body dto.UpdateStatusRequest true "New status"
// @Success 200 {object} response.ResponseStruct
// @Failure 400 {object} response.ResponseStruct
// @Failure 401 {object} response.ResponseStruct
// @Failure 403 {object} response.ResponseStruct
// @Failure 404 {object} response.ResponseStruct
// @Failure 500 {object} response.ResponseStruct
// @Security ApiKeyAuth
// @Router /api/v1/orders/{id}/status [patch]
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
			metrics.OrderStatusUpdatedTotal.WithLabelValues(string(req.Status), "not_found").Inc()
			return response.Reponse(c, 404, constant.NOT_FOUND, nil)
		}
		if errors.Is(err, errs.ORDER_STATUS_TRANSITION_INVALID) {
			return response.Reponse(c, 400, constant.INVALID_STATUS, nil)
		}
		metrics.OrderStatusUpdatedTotal.WithLabelValues(string(req.Status), "error").Inc()
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	metrics.OrderStatusUpdatedTotal.WithLabelValues(string(req.Status), "success").Inc()
	return response.Reponse(c, 200, constant.SUCCESS, nil)
}
