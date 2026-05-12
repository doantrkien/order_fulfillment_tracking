package handlers

import (
	"main/internal/services"

	"github.com/gofiber/fiber/v3"
)

type ReportHandler struct {
	reportService services.ReportService
}

func NewReportHandler(reportService services.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

func (h *ReportHandler) GetDailyReport(c fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"status":  200,
		"message": "success",
		"data":    nil,
	})
}
