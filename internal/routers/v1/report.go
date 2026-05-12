package routers

import (
	"main/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupReportRouter(app *fiber.App, reportHandler *handlers.ReportHandler) {
	report := app.Group("api/v1/reports")
	report.Get("/daily", reportHandler.GetDailyReport)
}
