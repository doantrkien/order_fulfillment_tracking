package models

import basemodel "main/internal/models/base_model"

type OrderEvent struct {
	basemodel.BaseModel
    EventID     string      `gorm:"column:event_id;uniqueIndex;not null" json:"event_id"`

    OrderID     int64       `gorm:"column:order_id;not null;index" json:"order_id"`
    Order       Order       `gorm:"foreignKey:OrderID" json:"-"`

    EventType   string      `gorm:"column:event_type;not null" json:"event_type"`
    NewStatus   OrderStatus `gorm:"column:new_status;not null" json:"new_status"`
    DriverID    *int64      `gorm:"column:driver_id" json:"driver_id,omitempty"`
    OccurredAt  time.Time   `gorm:"column:occurred_at" json:"occurred_at"`
    ProcessedAt time.Time   `gorm:"column:processed_at" json:"processed_at"`
}