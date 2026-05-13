package etl

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type DailyReport struct {
	Date            time.Time
	TotalOrders     int64
	TotalCreated    int64
	TotalDelivered  int64
	TotalCancelled  int64
	TotalRefunded   int64
	TotalIncome     float64
	AvgDeliverTime  float64
	StatusBreakdown map[string]int64 // Dùng map trong Go, sau đó parse thành JSONB
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
	query := `
		SELECT 
			COUNT(id) as total_orders,
			COUNT(CASE WHEN status = 'created' THEN 1 END) as total_created,
			COUNT(CASE WHEN status = 'delivered' THEN 1 END) as total_delivered,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as total_cancelled,
			COUNT(CASE WHEN status = 'refunded' THEN 1 END) as total_refunded,
			COUNT(CASE WHEN status = 'paid' THEN 1 END) as total_paid,
			COUNT(CASE WHEN status = 'packed' THEN 1 END) as total_packed,
			COALESCE(SUM(total_amount), 0) as total_income
		FROM orders 
		WHERE DATE(created_at) = $1
	`

	var report DailyReport
	report.Date = reportDate
	report.StatusBreakdown = make(map[string]int64)
	var totalPaid, totalPacked int64

	// Thực thi query và quét dữ liệu vào struct
	err := p.DB.QueryRowContext(ctx, query, dateStr).Scan(
		&report.TotalOrders,
		&report.TotalCreated,
		&report.TotalDelivered,
		&report.TotalCancelled,
		&report.TotalRefunded,
		&totalPaid,
		&totalPacked,
		&report.TotalIncome,
	)
	if err != nil {
		return fmt.Errorf("lỗi ở bước Extract/Transform (Orders): %v", err)
	}

	// Ghi nhận các trạng thái phụ vào map chuẩn bị cho JSONB
	report.StatusBreakdown["paid"] = totalPaid
	report.StatusBreakdown["packed"] = totalPacked

	// (Tuỳ chọn) Tính avg_deliver_time:
	// Em sẽ cần query thêm bảng order_events kết hợp với orders ở đây.
	// report.AvgDeliverTime = calculateAvgDelivery(p.DB, dateStr)

	// BƯỚC 3: LOAD
	return p.loadToReportsTable(ctx, report)
}

// loadToReportsTable thực hiện thao tác Upsert (Insert hoặc Update nếu đã tồn tại)
func (p *Processor) loadToReportsTable(ctx context.Context, report DailyReport) error {
	// Chuyển map từ Go thành mảng byte JSON để lưu vào cột JSONB
	statusBreakdownJSON, err := json.Marshal(report.StatusBreakdown)
	if err != nil {
		return fmt.Errorf("lỗi khi parse JSONB: %v", err)
	}

	upsertQuery := `
		INSERT INTO reports (
			date, total_orders, total_created, total_delivered, 
			total_cancelled, total_refunded, total_income, 
			avg_deliver_time, status_breakdown, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (date) DO UPDATE SET
			total_orders = EXCLUDED.total_orders,
			total_created = EXCLUDED.total_created,
			total_delivered = EXCLUDED.total_delivered,
			total_cancelled = EXCLUDED.total_cancelled,
			total_refunded = EXCLUDED.total_refunded,
			total_income = EXCLUDED.total_income,
			avg_deliver_time = EXCLUDED.avg_deliver_time,
			status_breakdown = EXCLUDED.status_breakdown,
			updated_at = NOW();
	`

	_, err = p.DB.ExecContext(ctx, upsertQuery,
		report.Date.Format("2006-01-02"),
		report.TotalOrders,
		report.TotalCreated,
		report.TotalDelivered,
		report.TotalCancelled,
		report.TotalRefunded,
		report.TotalIncome,
		report.AvgDeliverTime,
		statusBreakdownJSON,
	)

	if err != nil {
		return fmt.Errorf("lỗi ở bước Load (Upsert Reports): %v", err)
	}

	log.Println("Load dữ liệu vào bảng Reports thành công!")
	return nil
}
