package models

import "fmt"

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
