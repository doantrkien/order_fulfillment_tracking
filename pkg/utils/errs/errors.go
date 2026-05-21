package errs

import "errors"

var (
	ERR_NOT_FOUND                   = errors.New("resource not found")
	ORDER_STATUS_TRANSITION_INVALID = errors.New("order status transition is invalid")
)
