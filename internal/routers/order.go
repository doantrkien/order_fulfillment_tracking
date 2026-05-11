package routers

import (
	"main/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupOrderRouter(app *fiber.App, orderHandler *handlers.OrderHandler) {

	order := app.Group("/orders")
	order.Get("/", orderHandler.GetAllOrder)
	order.Get("/:id", orderHandler.GetOrderDetail)
}
