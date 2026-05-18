package dto

import (
	"main/internal/models"
	"time"
)

type OrderRequest struct {
	TotalAmount     int64            `json:"total_amount"`
	Username        string           `json:"username"`
	UserPhone       string           `json:"user_phone"`
	ShippingAddress string           `json:"shipping_address"`
	UserInfo        *models.UserInfo `json:"user_info,omitempty"`
}

type OrderReponse struct {
	ID              int64              `json:"id"`
	TotalAmount     int64              `json:"total_amount"`
	Username        string             `json:"username"`
	UserPhone       string             `json:"user_phone"`
	ShippingAddress string             `json:"shipping_address"`
	UserInfo        *models.UserInfo   `json:"user_info,omitempty"`
	Status          models.OrderStatus `json:"status"`
	Ordered_at      time.Time          `json:"ordered_at"`
}
