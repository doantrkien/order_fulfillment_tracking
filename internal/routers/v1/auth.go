package routers

import (
	"main/internal/handlers"
	middleware "main/internal/middlewares"

	"github.com/gofiber/fiber/v3"
)

func SetupAuthRouter(app *fiber.App, authHandler *handlers.AuthHandler) {
	// Tạo một group route riêng cho auth
	auth := app.Group("/api/v1/auth")

	// 1. Các API Công khai (Public) - Không cần truyền token
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)

	// 2. Các API Bảo mật (Protected) - Bắt buộc phải có token hợp lệ mới được gọi
	// Chú ý: Dùng middleware.Authenticate() giống như bạn đã làm ở file order.go
	auth.Post("/logout", middleware.Authenticate(), authHandler.Logout)
}
