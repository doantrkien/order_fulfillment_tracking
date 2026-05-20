package dto

import (
	"main/internal/models"
	"time"
)

type OrderRequest struct {
	TotalAmount     int64  `json:"total_amount" validate:"required,gt=0"`
	Username        string `json:"username" validate:"required,min=5,max=100"`
	UserPhone       string `json:"user_phone" validate:"required,min=10,max=15"`
	ShippingAddress string `json:"shipping_address" validate:"required,min=2,max=255"`
}

type UpdateStatusRequest struct {
	Status models.OrderStatus `json:"status" validate:"required"`
}

type OrderReponse struct {
	ID              int64              `json:"id"`
	TotalAmount     int64              `json:"total_amount"`
	Username        string             `json:"username"`
	UserPhone       string             `json:"user_phone"`
	ShippingAddress string             `json:"shipping_address"`
	Status          models.OrderStatus `json:"status"`
	Ordered_at      time.Time          `json:"ordered_at"`
}

type OrderQuery struct {
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	Status string `query:"status"`
	Date   string `query:"date"`
}
