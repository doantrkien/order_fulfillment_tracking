package routers

import (
	"main/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupAIRouter(app *fiber.App, aiHandler *handlers.AIHandler) {
	aiRouter := app.Group("/api/v1/ai")
	aiRouter.Post("orders/:id/exception-analysis", aiHandler.AnalyzeException)
	aiRouter.Get("orders/:id/exception-analysis", aiHandler.GetLatestAnalysis)
	aiRouter.Put("orders/orders/customer-update-draft", aiHandler.UpdateDraf)
}
