package repositories

import (
	"main/internal/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReportRepository interface {
	GetDailyReport(date time.Time) (*models.Report, error)
	BuildDailyReport(start, end time.Time) (*models.Report, error)
	SaveReport(report *models.Report) (*models.Report, error)
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
	var report models.Report

	if err := r.db.Where("date = ?", date.Format("2006-01-02")).First(&report).Error; err != nil {
		return nil, err
	}

	return &report, nil
}

func (r *reportRepository) BuildDailyReport(start, end time.Time) (*models.Report, error) {
	report := &models.Report{Date: start}

	if err := r.db.Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&report.TotalOrders).Error; err != nil {
		return nil, err
	}

	var statusCounts []struct {
		Status models.OrderStatus
		Count  int64
	}

	if err := r.db.Model(&models.Order{}).
		Select("current_status AS status, count(*) AS count").
		Where("created_at >= ? AND created_at < ?", start, end).
		Group("current_status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}

	for _, statusCount := range statusCounts {
		switch statusCount.Status {
		case models.ORDER_STATUS_CREATED:
			report.TotalNew = statusCount.Count
		case models.ORDER_STATUS_DELIVERED:
			report.TotalDelivered = statusCount.Count
		case models.ORDER_STATUS_CANCELLED:
			report.TotalCancelled = statusCount.Count
		case models.ORDER_STATUS_REFUNDED:
			report.TotalRefunded = statusCount.Count
		}
	}

	if err := r.db.Model(&models.Order{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("created_at >= ? AND created_at < ? AND current_status = ?", start, end, models.ORDER_STATUS_DELIVERED).
		Scan(&report.TotalIncome).Error; err != nil {
		return nil, err
	}

	var avgSeconds float64
	if err := r.db.Table("orders AS o").
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM oe.event_at - o.created_at)), 0)").
		Joins("JOIN order_events oe ON oe.order_id = o.id").
		Where("oe.new_status = ? AND oe.event_at >= ? AND oe.event_at < ?", models.ORDER_STATUS_DELIVERED, start, end).
		Row().Scan(&avgSeconds); err != nil {
		return nil, err
	}

	report.AvgDeliverTime = avgSeconds / 3600

	return report, nil
}

func (r *reportRepository) SaveReport(report *models.Report) (*models.Report, error) {
	if err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "date"}},
		UpdateAll: true,
	}).Create(report).Error; err != nil {
		return nil, err
	}

	return report, nil
}
