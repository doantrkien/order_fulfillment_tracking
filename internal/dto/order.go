package dto

import (
	"main/internal/models"
	"time"
)

type OrderRequest struct {
	TotalAmount     int64  `json:"total_amount" validate:"required,gt=0" example:"100000"`
	Username        string `json:"username" validate:"required" example:"Supper Man"`
	UserPhone       string `json:"user_phone" validate:"required" example:"10123456789"`
	ShippingAddress string `json:"shipping_address" validate:"required" example:"Viet Nam"`
}

type UpdateStatusRequest struct {
	Status   models.OrderStatus `json:"status" validate:"required" example:"delivered"`
	DriverID *int64             `json:"driver_id,omitempty" example:"42"`
}

type OrderReponse struct {
	ID              int64              `json:"id" example:"1"`
	TotalAmount     int64              `json:"total_amount" example:"100000"`
	Username        string             `json:"username" example:"Nguyen Tien Khoa"`
	UserPhone       string             `json:"user_phone" example:"0977605602"`
	ShippingAddress string             `json:"shipping_address" example:"123 Nguyen Hue"`
	Status          models.OrderStatus `json:"status" example:"paid"`
	Ordered_at      time.Time          `json:"ordered_at" example:"2026-05-01T00:00:00Z"`
}
type CreateOrderResponse struct {
	ID              int64              `json:"id" example:"1"`
	Status          models.OrderStatus `json:"status" example:"created"`
	TotalAmount     int64              `json:"total_amount" example:"250000"`
	ShippingAddress string             `json:"shipping_address" example:"123 Nguyen Hue"`
	CreatedAt       time.Time          `json:"created_at" example:"2026-05-01T00:00:00Z"`
	UpdatedAt       time.Time          `json:"updated_at" example:"2026-05-01T00:00:00Z"`
}

type OrderQuery struct {
	PageNumber int    `query:"page"`
	LimitItems int    `query:"limit"`
	Status     string `query:"status"`
	Date       string `query:"date"`
}
