package repositories

import (
	"main/internal/models"
	"time"

	"gorm.io/gorm"
)

type ReportRepository interface {
	GetDailyReport(date time.Time) (*models.Report, error)
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *reportRepository {
	return &reportRepository{
		db: db,
	}
}

func (r *reportRepository) GetDailyReport(date time.Time) (*models.Report, error) {
	// TODO
	return nil, nil
}
