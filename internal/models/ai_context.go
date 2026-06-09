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

// AIContext aggregates order + events + driver notes to feed into the LLM.
type AIContext struct {
	OrderID       int64       `json:"order_id"`
	CreatedAt     time.Time   `json:"created_at"`
	CurrentStatus OrderStatus `json:"current_status"`
	TotalAmount   int64       `json:"total_amount"`
	Events        []AIEvent   `json:"events"`
}
