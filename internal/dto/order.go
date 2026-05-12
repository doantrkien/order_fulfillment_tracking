package dto

import (
	"main/internal/models"
	"time"
)

type OrderRequest struct {
	CustomerID   int64   `json:"customer_id"`
	TotalAmount  float64 `json:"total_amount"`
	ShippingAddr string  `json:"shipping_addr"`
}

type OrderReponse struct {
	TotalAmount  float64            `json:"total_amount"`
	ShippingAddr string             `json:"shipping_addr"`
	Status       models.OrderStatus `json:"status"`
	Ordered_at   time.Time          `json:"ordered_at"`
}
