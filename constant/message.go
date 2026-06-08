package constant

type MessageInfo struct {
	Code    string
	Message string
}

var (
	INVALID_INPUT = MessageInfo{Code: "INVALID_INPUT", Message: "Invalid input data"}

	BATCH_TOO_LARGE = MessageInfo{Code: "BATCH_TOO_LARGE", Message: "Batch too large, max 50000 events"}

	INVALID_STATUS = MessageInfo{Code: "INVALID_STATUS", Message: "Invalid status transition"}

	INTERNAL_SERVER_ERROR = MessageInfo{Code: "INTERNAL_SERVER_ERROR", Message: "Internal server error"}

	NOT_FOUND = MessageInfo{Code: "NOT_FOUND", Message: "Resource not found"}

	INVALID_CREDENTAIL = MessageInfo{Code: "NOT_FOUND", Message: "Invalid email or password"}

	UN_AUTHENTICATED = MessageInfo{Code: "UN_AUTHENTICATED", Message: "Authentication required"}

	UN_AUTHORIZED = MessageInfo{Code: "UN_AUTHORIZED", Message: "Permission denied"}

	SUCCESS = MessageInfo{Code: "SUCCESS", Message: "Success"}
)
