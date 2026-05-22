package dto

import (
	"time"
)

type GetDailyReportRequest struct {
	Date string `json:"date" query:"date" example:"2026-05-18"`
}

type DailyReportResponse struct {
	Date           time.Time `json:"date" example:"2026-05-13T00:00:00Z"`
	TotalOrders    int64     `json:"total_orders" example:"120"`
	TotalCreated   int64     `json:"total_created" example:"100"`
	TotalDelivered int64     `json:"total_delivered" example:"100"`
	TotalCanceled  int64     `json:"total_canceled" example:"20"`
	TotalRefunded  int64     `json:"total_refunded" example:"10"`
	TotalIncome    float64   `json:"total_income" example:"199.99"`
	AvgDeliverTime float64   `json:"avg_deliver_time" example:"2.5"`
}
