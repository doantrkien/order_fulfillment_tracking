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

	ERR_GEMINI_API_KEY_EMPTY           = &AppError{Err: constant.GEMINI_API_KEY_EMPTY, StatusCode: 400}
	ERR_GEMINI_CLIENT_CREATE_FAILED    = &AppError{Err: constant.GEMINI_CLIENT_CREATE_FAILED, StatusCode: 500}
	ERR_GEMINI_CLIENT_TIMEOUT          = &AppError{Err: constant.GEMINI_CLIENT_TIMEOUT, StatusCode: 504}
	ERR_GEMINI_CLIENT_RETRY_LIMIT      = &AppError{Err: constant.GEMINI_CLIENT_RETRY_LIMIT, StatusCode: 503}
	ERR_GEMINI_GENERATE_CONTENT_FAILED = &AppError{Err: constant.GEMINI_GENERATE_CONTENT_FAILED, StatusCode: 500}
	ERR_AI_DISABLED                    = &AppError{Err: constant.AI_DISABLED, StatusCode: 503}
	ERR_AI_INPUT_TOO_LARGE             = &AppError{Err: constant.AI_INPUT_TOO_LARGE, StatusCode: 400}
	ERR_AI_PING_FAILED                 = &AppError{Err: constant.AI_PING_FAILED, StatusCode: 503}
	ERR_AI_RESPONSE_VALIDATION_FAILED  = &AppError{Err: constant.AI_RESPONSE_VALIDATION_FAILED, StatusCode: 500}		
)	
