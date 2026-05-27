package routers

import (
	"main/internal/handlers"
	"main/internal/middlewares"

	"github.com/gofiber/fiber/v3"
)

func SetupOrderRouter(app *fiber.App, orderHandler *handlers.OrderHandler) {
	order := app.Group("api/v1/orders",
		middlewares.Authenticate(),
	)

	order.Get("", middlewares.Authorize([]string{"admin", "driver"}), orderHandler.GetAllOrder)
	order.Get("/:id", middlewares.Authorize([]string{"admin", "driver"}), orderHandler.GetOrderDetail)
	order.Post("", middlewares.Authorize([]string{"admin"}), orderHandler.CreateOrder)
	order.Patch("/:id/status", middlewares.Authorize([]string{"admin", "driver"}), orderHandler.UpdateOrderStatus)
}
