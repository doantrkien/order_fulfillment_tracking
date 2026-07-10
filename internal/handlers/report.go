package handlers

import (
	"errors"
	"main/constant"
	"main/errs"
	dto "main/internal/dto/api"
	"main/internal/services"
	"main/response"
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
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 404 {object} response.ErrorNotFoundResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/reports/daily [get]
func (h *ReportHandler) GetDailyReport(c fiber.Ctx) error {
	dateStr := c.Query("date")
	if dateStr == "" {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	today := time.Now().Truncate(24 * time.Hour)
	if date.Truncate(24 * time.Hour).After(today) {
		return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
	}

	report, err := h.reportService.GetDailyReport(date)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) {
			return response.ResponseError(c, errs.ERR_NOT_FOUND, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, report)
}

// CreateDailyReport godoc
// @Summary Create daily report manually
// @Tags Report
// @Accept json
// @Produce json
// @Param request body dto.GetDailyReportRequest true "Date request"
// @Success 201 {object} response.ResponseStruct{data=dto.DailyReportResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security BearerAuth
// @Router /api/v1/reports/daily [post]
func (h *ReportHandler) CreateDailyReport(c fiber.Ctx) error {
	var req dto.GetDailyReportRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	if req.Date == "" {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	today := time.Now().Truncate(24 * time.Hour)
	if date.Truncate(24 * time.Hour).After(today) {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	report, err := h.reportService.CreateDailyReport(date)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, 201, constant.SUCCESS.Message, report)
}
