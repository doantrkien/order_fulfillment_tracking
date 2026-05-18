package dto

import (
	"main/internal/models"
	"time"
)

type OrderRequest struct {
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
	TotalAmount   int64  `json:"total_amount"`
	ShippingAddr  string `json:"shipping_addr"`
}

type OrderReponse struct {
	CustomerName  string             `json:"customer_name"`
	CustomerPhone string             `json:"customer_phone"`
	TotalAmount   int64              `json:"total_amount"`
	ShippingAddr  string             `json:"shipping_addr"`
	Status        models.OrderStatus `json:"status"`
	Ordered_at    time.Time          `json:"ordered_at"`
}

type OrderQuery struct {
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	Status string `query:"status"`
	Date   string `query:"date"`
}
