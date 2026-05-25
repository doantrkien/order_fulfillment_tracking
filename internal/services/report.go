package services

import (
	"errors"
	"main/internal/models"
	"main/internal/repositories"
	"time"

	"gorm.io/gorm"
)

type ReportService interface {
	GetDailyReport(date time.Time) (*models.Report, error)
	CreateDailyReport(date time.Time) (*models.Report, error)
}

type reportService struct {
	reportRepo repositories.ReportRepository
}

func NewReportService(reportRepo repositories.ReportRepository) ReportService {
	return &reportService{reportRepo: reportRepo}
}

func (s *reportService) GetDailyReport(date time.Time) (*models.Report, error) {
	report, err := s.reportRepo.GetDailyReport(date)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.CreateDailyReport(date)
		}
		return nil, err
	}
	return report, nil
}

func (s *reportService) CreateDailyReport(date time.Time) (*models.Report, error) {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		loc = time.UTC
	}

	periodStart := time.Date(date.Year(), date.Month(), date.Day(), 3, 0, 0, 0, loc)
	periodEnd := periodStart.Add(24 * time.Hour)

	report, err := s.reportRepo.BuildDailyReport(periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	return s.reportRepo.SaveReport(report)
}
