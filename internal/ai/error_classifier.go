package ai

import (
	"context"
	"errors"
	"net"
)

// ClassifyError maps an AI adapter error to a fallback reason string.
// This is used by the ExceptionAnalyzer to record why fallback was triggered.
func ClassifyError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return FallbackReasonTimeout
	}

	if errors.Is(err, context.Canceled) {
		return FallbackReasonConnectionError
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return FallbackReasonTimeout
	}

	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		return FallbackReasonConnectionError
	}

	// Default: treat unknown errors as connection errors
	return FallbackReasonConnectionError
}
