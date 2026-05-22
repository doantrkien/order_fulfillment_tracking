package handlers

import (
	"errors"
	"main/internal/dto"
	"main/internal/services"
	"main/pkg/utils/constant"
	"main/pkg/utils/errs"
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

// GetDailyReport godoc
// @Summary Daily order report
// @Tags Report
// @Accept json
// @Produce json
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} response.ResponseStruct{data=dto.DailyReportResponse}
// @Failure 400 {object} response.ResponseStruct
// @Failure 401 {object} response.ResponseStruct
// @Failure 403 {object} response.ResponseStruct
// @Failure 404 {object} response.ResponseStruct
// @Failure 500 {object} response.ResponseStruct
// @Security ApiKeyAuth
// @Router /api/v1/reports/daily [get]
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
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.Reponse(c, 404, constant.NOT_FOUND, nil)
		}
		return response.Reponse(c, 500, constant.ERROR, nil)
	}

	return response.Reponse(c, 200, constant.SUCCESS, report)
}

// CreateDailyReport godoc
// @Summary Create daily report manually
// @Tags Report
// @Accept json
// @Produce json
// @Param request body dto.GetDailyReportRequest true "Date request"
// @Success 201 {object} response.ResponseStruct{data=dto.DailyReportResponse}
// @Failure 400 {object} response.ResponseStruct
// @Failure 401 {object} response.ResponseStruct
// @Failure 403 {object} response.ResponseStruct
// @Failure 500 {object} response.ResponseStruct
// @Security ApiKeyAuth
// @Router /api/v1/reports/daily [post]
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
