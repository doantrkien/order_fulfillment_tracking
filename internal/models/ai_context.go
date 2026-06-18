package models

import "time"

type AIEvent struct {
	EventAt        time.Time   `json:"event_at"`
	PreviousStatus OrderStatus `json:"previous_status"`
	NewStatus      OrderStatus `json:"new_status"`
	DriverID       *int64      `json:"driver_id,omitempty"`
	DriverNote     *string     `json:"driver_note,omitempty"`
	UpdatedBy      string      `json:"updated_by"`
}

type AIContext struct {
	OrderID         int64       `json:"order_id"`
	CreatedAt       time.Time   `json:"created_at"`
	CurrentStatus   OrderStatus `json:"current_status"`
	TotalAmount     int64       `json:"total_amount"`
	CustomerName    string      `json:"customer_name"`
	ShippingAddress string      `json:"shipping_address"`
	PaymentStatus   string      `json:"payment_status"`
	RefundStatus    *string     `json:"refund_status,omitempty"`
	Events          []AIEvent   `json:"events"`
}

func DerivePaymentStatus(status OrderStatus) string {
	switch status {
	case ORDER_STATUS_CREATED:
		return "pending"
	case ORDER_STATUS_PAID, ORDER_STATUS_PACKED, ORDER_STATUS_SHIPPED, ORDER_STATUS_DELIVERED:
		return "paid"
	case ORDER_STATUS_REFUNDED:
		return "refunded"
	case ORDER_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "unknown"
	}
}

func DeriveRefundStatus(status OrderStatus) *string {
	if status == ORDER_STATUS_REFUNDED {
		s := "refunded"
		return &s
	}
	return nil
}
