package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFallbackTemplate(t *testing.T) {
	tests := []struct {
		exceptionType string
		wantContains  string
	}{
		{"STUCK_ORDER", "slower than expected"},
		{"DELIVERY_FAILURE", "address [REDACTED_SHIPPING_ADDRESS]"},
		{"INVALID_TRANSITION", "mismatch in your order status"},
		{"SKIPPED_STATUS", "unusual updates"},
		{"DUPLICATE_EVENT", "duplication has been recorded"},
		{"CANCELLATION_ANOMALY", "unusual cancellation request"},
		{"REFUND_ANOMALY", "refund request"},
		{"OTHER", "unexpected issue related to your order"},
		{"SOME_RANDOM_TYPE", "unforeseen issue"}, // Dẫn về DEFAULT
		{"", "unforeseen issue"},                 // Dẫn về DEFAULT
	}

	for _, tc := range tests {
		t.Run(tc.exceptionType, func(t *testing.T) {
			got := GetFallbackTemplate(tc.exceptionType)
			assert.Contains(t, got, "[REDACTED_CUSTOMER_NAME]")
			assert.Contains(t, got, tc.wantContains)
		})
	}
}
