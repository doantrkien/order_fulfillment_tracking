package routers

import (
	"main/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupAIRouter(app *fiber.App, aiHandler *handlers.AIHandler) {
	aiRouter := app.Group("/api/v1/ai")
	aiRouter.Post("orders/:id/exception-analysis", aiHandler.AnalyzeException)
	aiRouter.Get("orders/:id/insights/latest", aiHandler.GetLatestAnalysis)
	aiRouter.Post("orders/customer-update-draft", aiHandler.GenerateDraft)
	aiRouter.Post("evaluations/order-exceptions", aiHandler.TriggerEvaluation)
	aiRouter.Get("evaluations/:run_id", aiHandler.GetEvaluationRun)
	aiRouter.Get("evaluations/:run_id/details", aiHandler.GetEvaluationDetails)
}
