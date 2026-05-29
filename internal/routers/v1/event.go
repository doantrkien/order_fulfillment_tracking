package routers

import (
	"main/internal/handlers"
	middleware "main/internal/middlewares"
	// Giữ nguyên đường dẫn của bạn
	"github.com/gofiber/fiber/v3"
)

func SetupOrderEventRouter(app *fiber.App, orderEventHandler *handlers.OrderEventHandler) {
	orderEvent := app.Group("/api/v1/order-events", middleware.Authenticate())

	// Đã xóa bỏ []string{}, truyền trực tiếp "admin"
	orderEvent.Post("/import", middleware.Authorize("admin"), orderEventHandler.ImportOrderEvents)
}
