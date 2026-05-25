package errs

import "errors"

var (
	ERR_NOT_FOUND                       = errors.New("resource not found")
	ERR_ORDER_STATUS_TRANSITION_INVALID = errors.New("invalid order status transition")
)
