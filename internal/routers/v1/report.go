package routers

import (
	"main/internal/handlers"
	middleware "main/internal/middlewares"

	"github.com/gofiber/fiber/v3"
)

func SetupReportRouter(app *fiber.App, reportHandler *handlers.ReportHandler) {
	report := app.Group("api/v1/reports", middleware.Authenticate())

	// Đã xóa bỏ []string{}
	report.Get("/daily", middleware.Authorize("admin"), reportHandler.GetDailyReport)
	report.Post("/daily", middleware.Authorize("admin"), reportHandler.CreateDailyReport)
}
