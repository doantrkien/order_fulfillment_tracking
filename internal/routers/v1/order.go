package routers

import (
	"main/internal/handlers"
	middleware "main/internal/middlewares"

	"github.com/gofiber/fiber/v3"
)

func SetupOrderRouter(app *fiber.App, orderHandler *handlers.OrderHandler) {
	// Sửa middlewares thành middleware
	order := app.Group("api/v1/orders", middleware.Authenticate())

	// Đã xóa bỏ []string{}, truyền trực tiếp các role cách nhau bằng dấu phẩy
	order.Get("", middleware.Authorize("admin", "customer", "driver"), orderHandler.GetAllOrder)
	order.Get("/:id", middleware.Authorize("admin", "customer", "driver"), orderHandler.GetOrderDetail)
	order.Post("", middleware.Authorize("customer", "admin"), orderHandler.CreateOrder)
	order.Patch("/:id/status", middleware.Authorize("admin", "customer", "driver"), orderHandler.UpdateOrderStatus)
}
