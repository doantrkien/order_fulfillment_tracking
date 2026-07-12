package routers

import (
	"main/internal/handlers"
	"main/internal/middlewares"

	"github.com/gofiber/fiber/v3"
)

func SetupOrderEventRouter(app *fiber.App, orderEventHandler *handlers.OrderEventHandler) {
	orderEvent := app.Group("/api/v1/order-events", middlewares.Authenticate())
	orderEvent.Post("/import", middlewares.Authorize([]string{"admin"}), orderEventHandler.ImportOrderEvents)

	// Route riêng cho driver: chỉ cập nhật note, không tạo event hàng loạt
	orders := app.Group("/api/v1/orders", middlewares.Authenticate())
	orders.Patch("/:id/driver-note", middlewares.Authorize([]string{"driver", "admin"}), orderEventHandler.UpdateDriverNote)
}
