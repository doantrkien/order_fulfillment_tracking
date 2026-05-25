package dto

import (
	"main/internal/models"
	"time"
)

type OrderRequest struct {
	TotalAmount     int64  `json:"total_amount" validate:"required,gt=0" example:"100000"`
	Username        string `json:"username" validate:"required,not null" example:"Supper Man"`
	UserPhone       string `json:"user_phone" validate:"required,not null,length=10|length=11" example:"10123456789"`
	ShippingAddress string `json:"shipping_address" validate:"required,not null" example:"Viet Nam"`
}

type UpdateStatusRequest struct {
	Status models.OrderStatus `json:"status" validate:"required,not null" example:"delivered"`
}

type OrderReponse struct {
	ID              int64              `json:"id" example:"1"`
	TotalAmount     int64              `json:"total_amount" example:"100000"`
	Username        string             `json:"username" example:"Supper Man"`
	UserPhone       string             `json:"user_phone" example:"0123456789"`
	ShippingAddress string             `json:"shipping_address" example:"Viet Nam"`
	Status          models.OrderStatus `json:"status" example:"paid"`
	Ordered_at      time.Time          `json:"ordered_at" example:"2026-05-01T00:00:00Z"`
}

type OrderQuery struct {
	PageNumber int    `query:"page"`
	LimitItems int    `query:"limit"`
	Status     string `query:"status"`
	Date       string `query:"date"`
}
