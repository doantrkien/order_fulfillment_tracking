package routers

import (
	"main/internal/handlers"
	"main/internal/middlewares"

	"github.com/gofiber/fiber/v3"
)

func SetupReportRouter(app *fiber.App, reportHandler *handlers.ReportHandler) {
	report := app.Group("api/v1/reports", middlewares.Authenticate())
	report.Get("/daily", middlewares.Authorize([]string{"admin"}), reportHandler.GetDailyReport)
	report.Post("/daily", middlewares.Authorize([]string{"admin"}), reportHandler.CreateDailyReport)
}
