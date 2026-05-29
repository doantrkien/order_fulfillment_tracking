package routers

import (
	"main/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupAuthRouter(app *fiber.App, authHandler *handlers.AuthHandler) {
	app.Post("/api/v1/auth/login", authHandler.Login)
}
