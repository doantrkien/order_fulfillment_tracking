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
	return &reportRepository{db: db}
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

	type summaryResult struct {
		TotalOrders    int64
		TotalNew       int64
		TotalDelivered int64
		TotalCancelled int64
		TotalRefunded  int64
		TotalIncome    int64
	}

	var summary summaryResult

	err := r.db.Raw(`
		SELECT
			COUNT(id) AS total_orders,
			COUNT(CASE WHEN current_status = 'created' THEN 1 END) AS total_new,
			COUNT(CASE WHEN current_status = 'delivered' THEN 1 END) AS total_delivered,
			COUNT(CASE WHEN current_status = 'cancelled' THEN 1 END) AS total_cancelled,
			COUNT(CASE WHEN current_status = 'refunded' THEN 1 END) AS total_refunded,
			COALESCE(SUM(CASE WHEN current_status = 'delivered' THEN total_amount ELSE 0 END), 0) AS total_income
		FROM orders
		WHERE created_at >= ? AND created_at < ?
	`, start, end).Scan(&summary).Error
	if err != nil {
		return nil, err
	}

	report.TotalOrders = summary.TotalOrders
	report.TotalNew = summary.TotalNew
	report.TotalDelivered = summary.TotalDelivered
	report.TotalCancelled = summary.TotalCancelled
	report.TotalRefunded = summary.TotalRefunded
	report.TotalIncome = summary.TotalIncome

	var avgSeconds float64
	err = r.db.Raw(`
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM oe.event_at - o.created_at)), 0)
		FROM orders o
		JOIN order_events oe ON oe.order_id = o.id
		WHERE oe.new_status = 'delivered'
		  AND oe.event_at >= ? AND oe.event_at < ?
	`, start, end).Scan(&avgSeconds).Error
	if err != nil {
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
