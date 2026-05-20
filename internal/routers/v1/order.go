package routers

import (
	"main/internal/handlers"
	"main/internal/middlewares"

	"github.com/gofiber/fiber/v3"
)

func SetupOrderRouter(app *fiber.App, orderHandler *handlers.OrderHandler) {
	order := app.Group("api/v1/orders", middlewares.Authenticate())
	//order := app.Group("api/v1/orders")
	order.Get("", middlewares.Authorize([]string{"admin", "customer", "shipper"}), orderHandler.GetAllOrder)
	//order.Get("",  orderHandler.GetAllOrder)

	order.Get("/:id", middlewares.Authorize([]string{"admin", "customer", "shipper"}), orderHandler.GetOrderDetail)
	order.Post("", middlewares.Authorize([]string{"customer"}), orderHandler.CreateOrder)
	//order.Post("", orderHandler.CreateOrder)
	order.Patch("/:id/status", middlewares.Authorize([]string{"admin", "customer"}), orderHandler.UpdateOrderStatus)
}
