package services

import (
	"errors"
	"main/internal/models"
	"main/internal/repositories"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type ReportService interface {
	GetDailyReport(date time.Time) (*models.Report, error)
	CreateDailyReport(date time.Time) (*models.Report, error)
}

type reportService struct {
	reportRepo repositories.ReportRepository
	sf         singleflight.Group
}

func NewReportService(reportRepo repositories.ReportRepository) ReportService {
	return &reportService{reportRepo: reportRepo}
}

func (s *reportService) GetDailyReport(date time.Time) (*models.Report, error) {
	loc := time.FixedZone("Asia/Ho_Chi_Minh", 7*3600)
	now := time.Now().In(loc)
	reqDateStr := date.Format("2006-01-02")

	// Always generate dynamically if the requested date is today
	if reqDateStr == now.Format("2006-01-02") {
		v, err, _ := s.sf.Do(reqDateStr, func() (interface{}, error) {
			return s.CreateDailyReport(date)
		})
		if err != nil {
			return nil, err
		}
		return v.(*models.Report), nil
	}

	report, err := s.reportRepo.GetDailyReport(date)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			v, err, _ := s.sf.Do(reqDateStr, func() (interface{}, error) {
				return s.CreateDailyReport(date)
			})
			if err != nil {
				return nil, err
			}
			return v.(*models.Report), nil
		}
		return nil, err
	}
	return report, nil
}

func (s *reportService) CreateDailyReport(date time.Time) (*models.Report, error) {
	loc := time.FixedZone("Asia/Ho_Chi_Minh", 7*3600)
	periodStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
	periodEnd := periodStart.Add(24 * time.Hour)

	report, err := s.reportRepo.BuildDailyReport(periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	return s.reportRepo.SaveReport(report)
}
