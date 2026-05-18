package models

import "fmt"

// IsValidStatus checks if a status string is a known OrderStatus value.
// Used by the service layer for basic input validation.
func IsValidStatus(s OrderStatus) bool {
	switch s {
	case ORDER_STATUS_CREATED, ORDER_STATUS_PAID, ORDER_STATUS_PACKED,
		ORDER_STATUS_SHIPPED, ORDER_STATUS_DELIVERED, ORDER_STATUS_CANCELLED,
		ORDER_STATUS_REFUNDED:
		return true
	}
	return false
}

func IsValidTransition(prev, next OrderStatus) bool {
	allowed, ok := validTransitions[prev]
	if !ok {
		return false
	}

	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

func ValidateEvent(e *OrderEvent) error {
	if !IsValidTransition(e.PreviousStatus, e.NewStatus) {
		return fmt.Errorf("invalid transition: %s -> %s", e.PreviousStatus, e.NewStatus)
	}
	return nil
}
