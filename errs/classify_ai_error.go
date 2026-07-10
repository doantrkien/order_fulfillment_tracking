package errs

import (
	"context"
	"errors"
	"main/constant"
	"net"
)

func ClassifyError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return constant.FallbackReasonTimeout
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return constant.FallbackReasonTimeout
	}

	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		return constant.FallbackReasonConnectionError
	}

	// Default: treat unknown errors as connection errors
	return constant.FallbackReasonConnectionError
}
