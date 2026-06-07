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

	// ── Query 1: orders created in window ───────────────────────────────────
	type createdSummary struct {
		TotalOrders int64
		TotalNew    int64
	}
	var created createdSummary
	err := r.db.Raw(`
		SELECT
			COUNT(id)                                                    AS total_orders,
			COUNT(CASE WHEN current_status = 'created' THEN 1 END)      AS total_new
		FROM orders
		WHERE created_at >= ? AND created_at < ?
	`, start, end).Scan(&created).Error
	if err != nil {
		return nil, err
	}

	// ── Query 2: event-based stats (delivered, cancelled, refunded, income) ─
	type eventSummary struct {
		TotalDelivered int64
		TotalCancelled int64
		TotalRefunded  int64
		TotalIncome    int64
	}
	var events eventSummary
	err = r.db.Raw(`
		SELECT
			COUNT(CASE WHEN oe.new_status = 'delivered'  THEN 1 END)                                AS total_delivered,
			COUNT(CASE WHEN oe.new_status = 'cancelled'  THEN 1 END)                                AS total_cancelled,
			COUNT(CASE WHEN oe.new_status = 'refunded'   THEN 1 END)                                AS total_refunded,
			COALESCE(SUM(CASE WHEN oe.new_status = 'delivered' THEN o.total_amount ELSE 0 END), 0)  AS total_income
		FROM order_events oe
		JOIN orders o ON o.id = oe.order_id
		WHERE oe.event_at >= ? AND oe.event_at < ?
	`, start, end).Scan(&events).Error
	if err != nil {
		return nil, err
	}

	// ── Query 3: avg delivery time using only order_events ──────────────────
	// Self-join: match the 'created' event and 'delivered' event per order,
	// then subtract — no join to orders table needed.
	var avgSeconds float64
	err = r.db.Raw(`
		SELECT COALESCE(
			AVG(
				EXTRACT(EPOCH FROM delivered.event_at - created.event_at)
			), 0)
		FROM order_events created
		JOIN order_events delivered ON delivered.order_id = created.order_id
		WHERE created.new_status  = 'created'
		  AND delivered.new_status = 'delivered'
		  AND delivered.event_at >= ? AND delivered.event_at < ?
	`, start, end).Scan(&avgSeconds).Error
	if err != nil {
		return nil, err
	}

	report.TotalOrders = created.TotalOrders
	report.TotalNew = created.TotalNew
	report.TotalDelivered = events.TotalDelivered
	report.TotalCancelled = events.TotalCancelled
	report.TotalRefunded = events.TotalRefunded
	report.TotalIncome = events.TotalIncome
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
