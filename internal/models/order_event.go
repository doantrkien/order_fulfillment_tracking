package models

import "time"

type OrderEvent struct {
	ID             int64       `gorm:"primaryKey;column:id" json:"id"`
	OrderID        int64       `gorm:"column:order_id;not null;index" json:"order_id"`
	Order          Order       `gorm:"foreignKey:OrderID" json:"-"`
	PreviousStatus OrderStatus `gorm:"column:previous_status;not null" json:"previous_status"`
	NewStatus      OrderStatus `gorm:"column:new_status;not null" json:"new_status"`
	DriverID       *int64      `gorm:"column:driver_id" json:"driver_id"`
	DriverNote     *string     `gorm:"column:driver_note" json:"driver_note,omitempty"`
	UpdatedBy      string      `gorm:"column:updated_by;not null" json:"updated_by"`
	EventAt        time.Time   `gorm:"column:event_at;not null;default:CURRENT_TIMESTAMP;index" json:"event_at"`
	CreatedAt      time.Time   `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time   `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
