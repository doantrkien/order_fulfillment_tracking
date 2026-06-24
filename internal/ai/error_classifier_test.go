package ai

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		inputErr error
		expected string
	}{
		{
			name:     "DeadlineExceeded maps to timeout",
			inputErr: context.DeadlineExceeded,
			expected: FallbackReasonTimeout,
		},
		{
			name:     "Canceled maps to connection error",
			inputErr: context.Canceled,
			expected: FallbackReasonConnectionError,
		},
		{
			name:     "Generic net.Error Timeout maps to timeout",
			inputErr: mockNetError{timeout: true},
			expected: FallbackReasonTimeout,
		},
		{
			name:     "Net OpError maps to connection error",
			inputErr: &net.OpError{Err: errors.New("connection reset by peer")},
			expected: FallbackReasonConnectionError,
		},
		{
			name:     "Unknown error maps to connection error",
			inputErr: errors.New("some weird parsing error"),
			expected: FallbackReasonConnectionError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ClassifyError(tc.inputErr)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// mockNetError implements net.Error interface
type mockNetError struct {
	timeout   bool
	temporary bool
}

func (e mockNetError) Error() string   { return "mock net error" }
func (e mockNetError) Timeout() bool   { return e.timeout }
func (e mockNetError) Temporary() bool { return e.temporary }
