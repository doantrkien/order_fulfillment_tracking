package models_test

import (
	"testing"

	"main/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     models.OrderStatus
		to       models.OrderStatus
		expected bool
	}{
		// Valid transitions: created →
		{"created → paid", models.ORDER_STATUS_CREATED, models.ORDER_STATUS_PAID, true},
		{"created → cancelled", models.ORDER_STATUS_CREATED, models.ORDER_STATUS_CANCELLED, true},

		// Valid transitions: paid →
		{"paid → packed", models.ORDER_STATUS_PAID, models.ORDER_STATUS_PACKED, true},
		{"paid → refunded", models.ORDER_STATUS_PAID, models.ORDER_STATUS_REFUNDED, true},

		// Valid transitions: packed →
		{"packed → shipped", models.ORDER_STATUS_PACKED, models.ORDER_STATUS_SHIPPED, true},

		// Valid transitions: shipped →
		{"shipped → delivered", models.ORDER_STATUS_SHIPPED, models.ORDER_STATUS_DELIVERED, true},

		// Invalid transitions: skipping steps
		{"created → delivered (skip)", models.ORDER_STATUS_CREATED, models.ORDER_STATUS_DELIVERED, false},
		{"created → shipped (skip)", models.ORDER_STATUS_CREATED, models.ORDER_STATUS_SHIPPED, false},
		{"paid → delivered (skip)", models.ORDER_STATUS_PAID, models.ORDER_STATUS_DELIVERED, false},

		// Invalid transitions: going backwards
		{"delivered → created (backward)", models.ORDER_STATUS_DELIVERED, models.ORDER_STATUS_CREATED, false},
		{"shipped → packed (backward)", models.ORDER_STATUS_SHIPPED, models.ORDER_STATUS_PACKED, false},

		// Invalid transitions: from terminal states
		{"delivered → anything", models.ORDER_STATUS_DELIVERED, models.ORDER_STATUS_SHIPPED, false},
		{"cancelled → anything", models.ORDER_STATUS_CANCELLED, models.ORDER_STATUS_PAID, false},
		{"refunded → anything", models.ORDER_STATUS_REFUNDED, models.ORDER_STATUS_CREATED, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsValidTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   models.OrderStatus
		expected bool
	}{
		// Valid statuses
		{"created", models.ORDER_STATUS_CREATED, true},
		{"paid", models.ORDER_STATUS_PAID, true},
		{"packed", models.ORDER_STATUS_PACKED, true},
		{"shipped", models.ORDER_STATUS_SHIPPED, true},
		{"delivered", models.ORDER_STATUS_DELIVERED, true},
		{"cancelled", models.ORDER_STATUS_CANCELLED, true},
		{"refunded", models.ORDER_STATUS_REFUNDED, true},

		// Invalid statuses
		{"empty string", "", false},
		{"random string", "random", false},
		{"typo", "cretaed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsValidStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateEvent(t *testing.T) {
	t.Run("Valid transition", func(t *testing.T) {
		e := &models.OrderEvent{
			PreviousStatus: models.ORDER_STATUS_CREATED,
			NewStatus:      models.ORDER_STATUS_PAID,
		}
		err := models.ValidateEvent(e)
		assert.NoError(t, err)
	})

	t.Run("Invalid transition", func(t *testing.T) {
		e := &models.OrderEvent{
			PreviousStatus: models.ORDER_STATUS_DELIVERED,
			NewStatus:      models.ORDER_STATUS_CREATED,
		}
		err := models.ValidateEvent(e)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid transition")
	})
}
