package routers

import (
	"main/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupOrderEventRouter(app *fiber.App, orderEventHandler *handlers.OrderEventHandler) {
	orderEvent := app.Group("/api/v1/order-events")
	orderEvent.Post("/", orderEventHandler.ImportOrderEvents)
}
