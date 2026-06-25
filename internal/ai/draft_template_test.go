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
		{"DELIVERY_FAILURE", "order to 123 Main St"},
		{"INVALID_TRANSITION", "status mismatch during your order"},
		{"SKIPPED_STATUS", "unusual update"},
		{"DUPLICATE_EVENT", "recorded a duplicate"},
		{"OTHER", "issue that has arisen regarding your order"},
		{"SOME_RANDOM_TYPE", "unexpected issue"}, // Dẫn về DEFAULT
		{"", "unexpected issue"},                 // Dẫn về DEFAULT
	}

	for _, tc := range tests {
		t.Run(tc.exceptionType, func(t *testing.T) {
			got := GetFallbackTemplate(tc.exceptionType, "John Doe", "123 Main St")
			assert.Contains(t, got, "John Doe")
			assert.Contains(t, got, tc.wantContains)
		})
	}
}
