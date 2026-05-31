package errs

import "main/constant"

type AppError struct {
	Err        constant.MessageInfo
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Err.Message
}

var (
	ERR_NOT_FOUND          = &AppError{Err: constant.NOT_FOUND, StatusCode: 404}
	ERR_INVALID_INPUT      = &AppError{Err: constant.INVALID_INPUT, StatusCode: 400}
	ERR_INVALID_STATUS     = &AppError{Err: constant.INVALID_STATUS, StatusCode: 400}
	ERR_INTERNAL_SERVER    = &AppError{Err: constant.INTERNAL_SERVER_ERROR, StatusCode: 500}
	ERR_UNAUTHENTICATED    = &AppError{Err: constant.UN_AUTHENTICATED, StatusCode: 401}
	ERR_UNAUTHORIZED       = &AppError{Err: constant.UN_AUTHORIZED, StatusCode: 403}
	ERR_BATCH_TOO_LARGE    = &AppError{Err: constant.BATCH_TOO_LARGE, StatusCode: 400}
	ERR_INVALID_CREDENTAIL = &AppError{Err: constant.INVALID_CREDENTAIL, StatusCode: 401}
)
