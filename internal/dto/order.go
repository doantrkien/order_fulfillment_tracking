package dto

import (
	"main/internal/models"
	"time"
)

type OrderRequest struct {
}

type OrderReponse struct {
	TotalAmount  float64            `json:"total_amount"`
	ShippingAddr string             `json:"shipping_addr"`
	Status       models.OrderStatus `json:"status"`
	Ordered_at   time.Time          `json:"ordered_at"`
}
