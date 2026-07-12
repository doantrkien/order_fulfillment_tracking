package handlers

import (
	"context"
	"main/constant"
	"main/errs"
	"main/internal/dto"
	"main/internal/services"
	"main/response"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type OrderEventHandler struct {
	orderEventService services.OrderEventService
}

func NewOrderEventHandler(orderEventService services.OrderEventService) *OrderEventHandler {
	return &OrderEventHandler{orderEventService: orderEventService}
}

// ImportOrderEvents godoc
// @Summary Import batch order events
// @Description Process a batch of order status update events concurrently. Each event is validated and processed in its own DB transaction. The response always returns aggregated counts, even on partial failure.
// @Tags Order Event
// @Accept json
// @Produce json
// @Param request body []dto.ImportOrderEventRequest true "List of events"
// @Success 200 {object} response.ResponseStruct{data=dto.ImportOrderEventsResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse{data=dto.ImportOrderEventsResponse}
// @Security BearerAuth
// @Router /api/v1/order-events/import [post]
func (h *OrderEventHandler) ImportOrderEvents(c fiber.Ctx) error {
	var req []dto.ImportOrderEventRequest

	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}
	const maxBatchSize = 50000
	if len(req) > maxBatchSize {
		return response.ResponseError(c, errs.ERR_BATCH_TOO_LARGE, nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	result, err := h.orderEventService.ImportOrderEvents(ctx, req)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, result)
	}
	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, result)
}

// UpdateDriverNote godoc
// @Summary Driver updates a note on an order
// @Description Allows a driver to add or update a note on the latest event of an order. Does NOT create a new event or change the order status.
// @Tags Order Event
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param request body dto.UpdateDriverNoteRequest true "Driver note"
// @Success 200 {object} response.ResponseStruct
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 404 {object} response.ErrorNotFoundResponse
// @Security BearerAuth
// @Router /api/v1/orders/{id}/driver-note [patch]
func (h *OrderEventHandler) UpdateDriverNote(c fiber.Ctx) error {
	orderID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || orderID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	var req dto.UpdateDriverNoteRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}
	if strings.TrimSpace(req.Note) == "" {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if err := h.orderEventService.UpdateDriverNote(ctx, orderID, req.Note); err != nil {
		return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
	}
	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, nil)
}

// DriverUpdateStatus godoc
// @Summary Driver updates order status (packed -> shipped only)
// @Description Allows a driver to update the status of an order. The order MUST currently be in 'packed' status. Driver can only set it to 'shipped'.
// @Tags Order Event
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param request body dto.DriverUpdateStatusRequest true "New status"
// @Success 200 {object} response.ResponseStruct
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Security BearerAuth
// @Router /api/v1/orders/{id}/status [patch]
func (h *OrderEventHandler) DriverUpdateStatus(c fiber.Ctx) error {
	orderID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || orderID <= 0 {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	var req dto.DriverUpdateStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}
	if strings.TrimSpace(req.Status) == "" {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	updatedBy, _ := c.Locals("email").(string)
	if updatedBy == "" {
		updatedBy = "driver"
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if err := h.orderEventService.DriverUpdateStatus(ctx, orderID, req.Status, updatedBy); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}
	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, nil)
}
