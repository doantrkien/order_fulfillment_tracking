package handlers

import (
	"main/internal/dto"
	"main/internal/services"
	"main/pkg/utils/constant"
	"main/pkg/utils/response"
	"time"

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
	dateStr := c.Query("date")
	if dateStr == "" {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	report, err := h.reportService.GetDailyReport(date)
	if err != nil {
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	return response.Reponse(c, 200, constant.SUCCESS, report)
}

func (h *ReportHandler) CreateDailyReport(c fiber.Ctx) error {
	var req dto.GetDailyReportRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	if req.Date == "" {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return response.Reponse(c, 400, constant.INVALID_INPUT, nil)
	}

	report, err := h.reportService.CreateDailyReport(date)
	if err != nil {
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	return response.Reponse(c, 201, constant.SUCCESS, report)
}
