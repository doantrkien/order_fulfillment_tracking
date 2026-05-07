package models

import (
	basemodel "main/internal/models/base_model"
	"time"

	"gorm.io/datatypes"
)

type Report struct {
	basemodel.BaseModel
	Date            time.Time      `gorm:"column:date;type:date;uniqueIndex" json:"date"`
	TotalOrders     int64          `gorm:"column:total_orders;default:0" json:"total_orders"`
	TotalCreated    int64          `gorm:"column:total_created;default:0" json:"total_created"`
	TotalDelivered  int64          `gorm:"column:total_delivered;default:0" json:"total_delivered"`
	TotalCanceled   int64          `gorm:"column:total_canceled;default:0" json:"total_canceled"`
	TotalRefunded   int64          `gorm:"column:total_refunded;default:0" json:"total_refunded"`
	TotalIncome     float64        `gorm:"column:total_income;type:decimal(15,2);default:0" json:"total_income"`
	AvgDeliverTime  float64        `gorm:"column:avg_deliver_time;type:decimal(10,2);default:0" json:"avg_deliver_time"`
	StatusBreakdown datatypes.JSON `gorm:"column:status_breakdown;type:jsonb" json:"status_breakdown"`
}
