package models

import "fmt"

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

// driverAllowedStatuses defines which statuses a driver is permitted to set.
// Financial/administrative statuses (cancelled, refunded) are admin-only.
var driverAllowedStatuses = map[OrderStatus]bool{
	ORDER_STATUS_SHIPPED:   true,
	ORDER_STATUS_DELIVERED: true,
}

func IsDriverAllowedStatus(status OrderStatus) bool {
	return driverAllowedStatuses[status]
}
