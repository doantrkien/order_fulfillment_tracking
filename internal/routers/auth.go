package routers

import (
	"main/internal/handlers"
	"main/internal/swagger"

	"github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"
)

func SetupRouter(app *fiber.App) {
	app.Use(swaggerui.New(swaggerui.Config{
		BasePath:    "/",
		FileContent: swagger.Spec,
		Path:        "swagger",
		Title:       "Order Fulfillment Tracking API",
	}))

	authen := app.Group("/auth")

	authen.Get("/login", handlers.Login)
}
