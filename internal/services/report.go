package services

import (
	"main/internal/models"
	"main/internal/repositories"
	"time"
)

type ReportService interface {
	GetDailyReport(date time.Time) (*models.Report, error)
}

type reportService struct {
	reportRepo repositories.ReportRepository
}

func NewReportService(reportRepo repositories.ReportRepository) ReportService {
	return &reportService{
		reportRepo: reportRepo,
	}
}

func (s *reportService) GetDailyReport(date time.Time) (*models.Report, error) {
	// TODO
	return nil, nil
}
