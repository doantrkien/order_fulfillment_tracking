package models

import (
	basemodel "main/internal/models/base_model"
)

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusPacked    OrderStatus = "packed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "canceled"
	OrderStatusRefunded  OrderStatus = "refunded"
)
var validTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusCreated:   {OrderStatusPaid, OrderStatusCancelled},
	OrderStatusPaid:      {OrderStatusPacked, OrderStatusRefunded},
	OrderStatusPacked:    {OrderStatusShipped},
	OrderStatusShipped:   {OrderStatusDelivered},
}
type Order struct {
	basemodel.BaseModel
	CustomerID   int64       `gorm:"column:customer_id;not null" json:"customer_id"`
	TotalAmount  float64     `gorm:"column:total_amount;type:decimal(12,2);not null" json:"total_amount"`
	ShippingAddr string      `gorm:"column:shipping_addr;not null" json:"shipping_addr"`
	Status       OrderStatus `gorm:"column:status;type:varchar(20);not null;default:'created'" json:"status"`
}
