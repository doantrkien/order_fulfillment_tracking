package constant

type MessageInfo struct {
	Code    string
	Message string
}

var (
	INVALID_INPUT = MessageInfo{Code: "INVALID_INPUT", Message: "Invalid input data"}

	INVALID_STATUS = MessageInfo{Code: "INVALID_STATUS", Message: "Invalid status transition"}

	INTERNAL_SERVER_ERROR = MessageInfo{Code: "INTERNAL_SERVER_ERROR", Message: "Internal server error"}

	NOT_FOUND = MessageInfo{Code: "NOT_FOUND", Message: "Resource not found"}

	UN_AUTHENTICATED = MessageInfo{Code: "UN_AUTHENTICATED", Message: "Authentication required"}

	UN_AUTHORIZED = MessageInfo{Code: "UN_AUTHORIZED", Message: "Permission denied"}

	ROUTE_NOT_FOUND = MessageInfo{Code: "ROUTE_NOT_FOUND", Message: "Route not found"}

	SUCCESS = MessageInfo{Code: "SUCCESS", Message: "Success"}
)
