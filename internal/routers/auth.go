package routers

import (
	"main/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupRouter(app *fiber.App) {
	authen := app.Group("/auth")

	authen.Get("/login", handlers.Login)
}
