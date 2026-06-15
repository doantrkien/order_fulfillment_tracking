package models

import (
	"time"

	"gorm.io/datatypes"
)

type OrderStatus string

const (
	ORDER_STATUS_CREATED   OrderStatus = "created"
	ORDER_STATUS_PAID      OrderStatus = "paid"
	ORDER_STATUS_PACKED    OrderStatus = "packed"
	ORDER_STATUS_SHIPPED   OrderStatus = "shipped"
	ORDER_STATUS_DELIVERED OrderStatus = "delivered"
	ORDER_STATUS_CANCELLED OrderStatus = "cancelled"
	ORDER_STATUS_REFUNDED  OrderStatus = "refunded"
)

var validTransitions = map[OrderStatus][]OrderStatus{
	"":                     {ORDER_STATUS_CREATED},
	ORDER_STATUS_CREATED: {ORDER_STATUS_PAID, ORDER_STATUS_CANCELLED},
	ORDER_STATUS_PAID:    {ORDER_STATUS_PACKED, ORDER_STATUS_REFUNDED},
	ORDER_STATUS_PACKED:  {ORDER_STATUS_SHIPPED},
	ORDER_STATUS_SHIPPED: {ORDER_STATUS_DELIVERED},
}

type UserInfo struct {
	Username        string `json:"username"`
	UserPhone       string `json:"user_phone"`
	ShippingAddress string `json:"shipping_address"`
}

type Order struct {
	ID            int64          `gorm:"primaryKey;column:id" json:"id"`
	UserInfo      datatypes.JSON `gorm:"column:user_info;type:jsonb" json:"user_info"`
	TotalAmount   int64          `gorm:"column:total_amount;type:bigint;not null" json:"total_amount"`
	CurrentStatus OrderStatus    `gorm:"column:current_status;type:varchar(20);not null;default:'created';comment:Concurrency Lock" json:"current_status"`
	CreatedAt     time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
