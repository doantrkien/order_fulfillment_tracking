package dto

import (
	"time"
)

type GetDailyReportRequest struct {
	Date string `json:"date" query:"date"`
}

type DailyReportResponse struct {
	Date           time.Time `json:"date"`
	TotalOrders    int64     `json:"total_orders"`
	TotalCreated   int64     `json:"total_created"`
	TotalDelivered int64     `json:"total_delivered"`
	TotalCanceled  int64     `json:"total_canceled"`
	TotalRefunded  int64     `json:"total_refunded"`
	TotalIncome    float64   `json:"total_income"`
	AvgDeliverTime float64   `json:"avg_deliver_time"`
}
