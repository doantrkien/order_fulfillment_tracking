package models

import "time"

// AIEvent is a compact representation of an order event used to build prompts.
type AIEvent struct {
	EventAt        time.Time   `json:"event_at"`
	PreviousStatus OrderStatus `json:"previous_status"`
	NewStatus      OrderStatus `json:"new_status"`
	DriverID       *int64      `json:"driver_id,omitempty"`
	DriverNote     *string     `json:"driver_note,omitempty"`
	UpdatedBy      string      `json:"updated_by"`
}

// AIContext aggregates order + events into a structured input for the LLM.
// PaymentStatus và RefundStatus được derive từ CurrentStatus (Order model không có
// field riêng) để cung cấp ngữ cảnh rõ ràng hơn cho prompt.
type AIContext struct {
	OrderID       int64       `json:"order_id"`
	CreatedAt     time.Time   `json:"created_at"`
	CurrentStatus OrderStatus `json:"current_status"`
	TotalAmount     int64       `json:"total_amount"`
	CustomerName    string      `json:"customer_name"`
	ShippingAddress string      `json:"shipping_address"`
	PaymentStatus   string      `json:"payment_status"`          // derived: "paid" | "pending" | "refunded"
	RefundStatus    *string     `json:"refund_status,omitempty"` // non-nil chỉ khi status = refunded
	Events          []AIEvent   `json:"events"`
}

// DerivePaymentStatus suy ra trạng thái thanh toán từ CurrentStatus.
// Order.CurrentStatus là source of truth duy nhất — không có field payment riêng.
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

// DeriveRefundStatus trả về non-nil chỉ khi đơn hàng đang ở trạng thái refunded.
func DeriveRefundStatus(status OrderStatus) *string {
	if status == ORDER_STATUS_REFUNDED {
		s := "refunded"
		return &s
	}
	return nil
}
