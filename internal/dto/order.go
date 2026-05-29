package dto

import (
	"main/internal/models"
	"time"
)

type OrderRequest struct {
	UserID      int64 `json:"user_id" validate:"required,gt=0"`
	TotalAmount int64 `json:"total_amount" validate:"required,gt=0"`
}

type UpdateStatusRequest struct {
	Status models.OrderStatus `json:"status" validate:"required" example:"delivered"`
}

type OrderReponse struct {
	ID              int64              `json:"id"`
	UserID          int64              `json:"user_id"`
	Username        string             `json:"username"`
	UserPhone       string             `json:"user_phone"`
	ShippingAddress string             `json:"shipping_address"`
	TotalAmount     int64              `json:"total_amount"`
	Status          models.OrderStatus `json:"status"`
	Ordered_at      time.Time          `json:"ordered_at"`
}

type OrderQuery struct {
	PageNumber int    `query:"page"`
	LimitItems int    `query:"limit"`
	Status     string `query:"status"`
	Date       string `query:"date"`
}
