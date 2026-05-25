package handlers

import (
	"context"
	"main/internal/dto"
	"main/internal/services"
	"main/pkg/utils/response"
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
// @Failure 400 {object} response.ResponseStruct
// @Failure 500 {object} response.ResponseStruct{data=dto.ImportOrderEventsResponse}
// @Security ApiKeyAuth
// @Router /api/v1/order-events/import [post]
// func (h *OrderEventHandler) ImportOrderEvents(c fiber.Ctx) error {
func (h *OrderEventHandler) ImportOrderEvents(c fiber.Ctx) error {
	var req []dto.ImportOrderEventRequest

	if err := c.Bind().Body(&req); err != nil {
		return response.Reponse(c, 400, "invalid payload", nil)
	}

	const maxBatchSize = 1000
	if len(req) > maxBatchSize {
		return response.Reponse(c, 400, "batch too large, max 1000 events", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	result, err := h.orderEventService.ImportOrderEvents(ctx, req)
	if err != nil {
		return response.Reponse(c, 500, err.Error(), result)
	}
	return response.Reponse(c, 200, "batch processed", result)
}
