package etl

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type DailyReport struct {
	Date           time.Time
	TotalOrders    int64
	TotalNew       int64
	TotalDelivered int64
	TotalCancelled int64
	TotalRefunded  int64
	TotalIncome    float64
	AvgDeliverTime float64
}

type Processor struct {
	DB *sql.DB
}

// RunDailyReport thực thi tiến trình ETL cho một ngày cụ thể
func (p *Processor) RunDailyReport(ctx context.Context, reportDate time.Time) error {
	dateStr := reportDate.Format("2006-01-02")
	log.Printf("Bắt đầu job ETL Report cho ngày: %s\n", dateStr)

	// BƯỚC 1 & 2: EXTRACT & TRANSFORM
	// Tận dụng SQL để đếm và tính tổng hiệu quả hơn việc lặp hàng ngàn row trong Go
	periodEnd := time.Date(reportDate.Year(), reportDate.Month(), reportDate.Day(), 3, 0, 0, 0, reportDate.Location())
	periodStart := periodEnd.Add(-24 * time.Hour)

	query := `
		SELECT
			COUNT(id) as total_orders,
			COUNT(CASE WHEN current_status = 'created' THEN 1 END) as total_new,
			COUNT(CASE WHEN current_status = 'delivered' THEN 1 END) as total_delivered,
			COUNT(CASE WHEN current_status = 'cancelled' THEN 1 END) as total_cancelled,
			COUNT(CASE WHEN current_status = 'refunded' THEN 1 END) as total_refunded,
			COALESCE(SUM(total_amount), 0) as total_income
		FROM orders
		WHERE created_at >= $1 AND created_at < $2
	`

	var report DailyReport
	report.Date = reportDate

	// Thực thi query và quét dữ liệu vào struct
	err := p.DB.QueryRowContext(ctx, query, periodStart, periodEnd).Scan(
		&report.TotalOrders,
		&report.TotalNew,
		&report.TotalDelivered,
		&report.TotalCancelled,
		&report.TotalRefunded,
		&report.TotalIncome,
	)
	if err != nil {
		return fmt.Errorf("lỗi ở bước Extract/Transform (Orders): %v", err)
	}

	// Tính avg_deliver_time dựa trên order_events và orders
	avgQuery := `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM oe.event_at - o.created_at)), 0)
		FROM orders o
		JOIN order_events oe ON oe.order_id = o.id
		WHERE oe.new_status = 'delivered' AND oe.event_at >= $1 AND oe.event_at < $2
	`
	var avgSeconds float64
	if err := p.DB.QueryRowContext(ctx, avgQuery, periodStart, periodEnd).Scan(&avgSeconds); err != nil {
		return fmt.Errorf("lỗi ở bước Extract/Transform (AvgDeliverTime): %v", err)
	}
	report.AvgDeliverTime = avgSeconds / 3600

	// BƯỚC 3: LOAD
	return p.loadToReportsTable(ctx, report)
}

// loadToReportsTable thực hiện thao tác Upsert (Insert hoặc Update nếu đã tồn tại)
func (p *Processor) loadToReportsTable(ctx context.Context, report DailyReport) error {
	upsertQuery := `
		INSERT INTO reports (
			date, total_orders, total_new, total_delivered,
			total_cancelled, total_refunded, total_income,
			avg_deliver_time, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (date) DO UPDATE SET
			total_orders = EXCLUDED.total_orders,
			total_new = EXCLUDED.total_new,
			total_delivered = EXCLUDED.total_delivered,
			total_cancelled = EXCLUDED.total_cancelled,
			total_refunded = EXCLUDED.total_refunded,
			total_income = EXCLUDED.total_income,
			avg_deliver_time = EXCLUDED.avg_deliver_time,
			updated_at = NOW();
	`

	_, err := p.DB.ExecContext(ctx, upsertQuery,
		report.Date.Format("2006-01-02"),
		report.TotalOrders,
		report.TotalNew,
		report.TotalDelivered,
		report.TotalCancelled,
		report.TotalRefunded,
		report.TotalIncome,
		report.AvgDeliverTime,
	)

	if err != nil {
		return fmt.Errorf("lỗi ở bước Load (Upsert Reports): %v", err)
	}

	log.Println("Load dữ liệu vào bảng Reports thành công!")
	return nil
}
