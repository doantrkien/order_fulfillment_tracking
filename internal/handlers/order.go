package handlers

import (
	"errors"
	"fmt"
	"main/constant"
	"main/errs"
	"main/internal/dto"
	"main/internal/services"
	"main/response"
	"strconv"

	"github.com/go-playground/validator"
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
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/orders [get]
func (h *OrderHandler) GetAllOrder(c fiber.Ctx) error {
	var query dto.OrderQuery

	if err := c.Bind().Query(&query); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	if query.PageNumber <= 0 {
		query.PageNumber = 1
	}
	if query.LimitItems <= 0 {
		query.LimitItems = 10
	}
	if query.LimitItems > 100 {
		query.LimitItems = 100
	}

	result, totalItems, err := h.orderService.GetAllOrder(query)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.PaginatedSuccess(
		c,
		constant.SUCCESS.Message,
		result,
		query.PageNumber,
		query.LimitItems,
		totalItems,
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
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 404 {object} response.ErrorNotFoundResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrderDetail(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}
	fmt.Printf("GetOrderDetail: id=%d\n", id)
	order, err := h.orderService.GetOrder(id)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, order)
}

// CreateOrder godoc
// @Summary Create order
// @Description Create a new order
// @Tags Order
// @Accept json
// @Produce json
// @Param request body dto.OrderRequest true "Order details"
// @Success 201 {object} response.ResponseStruct
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
	var req dto.OrderRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	if err := validator.New().Struct(req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}
	_, err := h.orderService.CreateOrder(req)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 201, constant.SUCCESS.Message, nil)
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Tags Order
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param status body dto.UpdateStatusRequest true "New status"
// @Success 200 {object} response.ResponseStruct
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 404 {object} response.ErrorNotFoundResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}
	var req dto.UpdateStatusRequest

	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	if req.Status == "" {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	_, err = h.orderService.UpdateOrderStatus(id, string(req.Status))
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		if errors.Is(err, errs.ERR_INVALID_STATUS) {
			return response.ResponseError(c, errs.ERR_INVALID_STATUS, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}
	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, nil)
}
