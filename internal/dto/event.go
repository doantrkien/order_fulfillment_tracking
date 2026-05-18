package dto

import "time"

type ImportOrderEventRequest struct {
	OrderID   int64     `json:"order_id"`
	Status    string    `json:"status"`
	EventAt   time.Time `json:"event_at"`
	UpdatedBy string    `json:"updated_by"`
}

// EventError holds the detail of a single rejected or duplicate event.
type EventError struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}

type ImportOrderEventsResponse struct {
	Accepted  int          `json:"accepted_count"`
	Rejected  int          `json:"rejected_count"`
	Duplicate int          `json:"duplicate_count"`
	Errors    []EventError `json:"errors"`
}
