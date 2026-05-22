package dto

import "time"

type ImportOrderEventRequest struct {
	OrderID   int64     `json:"order_id" example:"100156"`
	Status    string    `json:"status" example:"packed"`
	EventAt   time.Time `json:"event_at" example:"2026-05-17T09:00:00Z"`
	UpdatedBy string    `json:"updated_by" example:"warehouse_staff_01"`
}

// EventError holds the detail of a single rejected or duplicate event.
type EventError struct {
	OrderID int64  `json:"order_id" example:"100157"`
	Status  string `json:"status" example:"delivered"`
	Reason  string `json:"reason" example:"Invalid transition from 'created' to 'delivered'"`
}

type ImportOrderEventsResponse struct {
	Accepted  int          `json:"accepted_count" example:"8"`
	Rejected  int          `json:"rejected_count" example:"1"`
	Duplicate int          `json:"duplicate_count" example:"1"`
	Errors    []EventError `json:"errors"`
}
