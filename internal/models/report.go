package models

import "time"

type Report struct {
	ID             int64     `gorm:"primaryKey;column:id" json:"id"`
	Date           time.Time `gorm:"column:date;type:date;uniqueIndex:reports_date_key" json:"date"`
	TotalOrders    int64     `gorm:"column:total_orders;default:0" json:"total_orders"`
	TotalNew       int64     `gorm:"column:total_new;default:0" json:"total_new"`
	TotalDelivered int64     `gorm:"column:total_delivered;default:0" json:"total_delivered"`
	TotalCancelled int64     `gorm:"column:total_cancelled;default:0" json:"total_cancelled"`
	TotalRefunded  int64     `gorm:"column:total_refunded;default:0" json:"total_refunded"`
	TotalIncome    float64   `gorm:"column:total_income;type:decimal(15,2);default:0" json:"total_income"`
	AvgDeliverTime float64   `gorm:"column:avg_deliver_time;type:decimal(10,2);default:0" json:"avg_deliver_time"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
