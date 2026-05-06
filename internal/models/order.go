package models

import basemodel "main/internal/models/base_model"
import "time"

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"
	OrderStatusPaid OrderStatus = "paid"
	OrderStatusPacked OrderStatus = "packed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCanceled  OrderStatus = "canceled"
	OrderStatusRefunded OrderStatus = "refunded"

)

type Order struct {
	basemodel.BaseModel
    CustomerID    int64       `gorm:"column:customer_id;not null" json:"customer_id"`
    CustomerEmail string      `gorm:"column:customer_email;not null" json:"customer_email"`
    TotalAmount   float64     `gorm:"column:total_amount;type:decimal(12,2);not null" json:"total_amount"`
    ShippingAddr  string      `gorm:"column:shipping_addr;not null" json:"shipping_addr"`
    Status        OrderStatus `gorm:"column:status;type:varchar(20);not null;default:'created'" json:"status"`
    PaidAt        *time.Time  `gorm:"column:paid_at" json:"paid_at,omitempty"`
    ShippedAt     *time.Time  `gorm:"column:shipped_at" json:"shipped_at,omitempty"`
    DeliveredAt   *time.Time  `gorm:"column:delivered_at" json:"delivered_at,omitempty"`
    CancelledAt   *time.Time  `gorm:"column:cancelled_at" json:"cancelled_at,omitempty"`
    RefundedAt    *time.Time  `gorm:"column:refunded_at" json:"refunded_at,omitempty"`
}	