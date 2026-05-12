package routers

import (
	"main/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupOrderRouter(app *fiber.App, orderHandler *handlers.OrderHandler) {
	order := app.Group("api/v1/orders")
	order.Get("", orderHandler.GetAllOrder)
	order.Get("/:id", orderHandler.GetOrderDetail)
	order.Post("", orderHandler.CreateOrder)
	order.Patch("/:id/status", orderHandler.UpdateOrderStatus)
}
