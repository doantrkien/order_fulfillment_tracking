package services

import (
	"main/internal/models"
	"main/internal/repositories"
	"time"
)

type ReportService interface {
	GetDailyReport(date time.Time) (*models.Report, error)
	CreateDailyReport(date time.Time) (*models.Report, error)
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
	return s.reportRepo.GetDailyReport(date)
}

func (s *reportService) CreateDailyReport(date time.Time) (*models.Report, error) {
	periodStart := time.Date(date.Year(), date.Month(), date.Day(), 3, 0, 0, 0, date.Location())
	periodEnd := periodStart.Add(24 * time.Hour)

	report, err := s.reportRepo.BuildDailyReport(periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	return s.reportRepo.SaveReport(report)
}
