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

	AI_API_KEY_EMPTY = MessageInfo{Code: "AI_API_KEY_EMPTY", Message: "AI API key is empty"}

	AI_CLIENT_CREATE_FAILED = MessageInfo{Code: "AI_CLIENT_CREATE_FAILED", Message: "Failed to create AI client"}

	AI_CLIENT_TIMEOUT = MessageInfo{Code: "AI_CLIENT_TIMEOUT", Message: "AI client timeout"}

	AI_CLIENT_RETRY_LIMIT = MessageInfo{Code: "AI_CLIENT_RETRY_LIMIT", Message: "AI client retry limit"}

	AI_GENERATE_CONTENT_FAILED = MessageInfo{Code: "AI_GENERATE_CONTENT_FAILED", Message: "AI generate content failed"}

	AI_RESPONSE_VALIDATION_FAILED = MessageInfo{Code: "AI_RESPONSE_VALIDATION_FAILED", Message: "AI response validation failed"}
	
	AI_DISABLED        = MessageInfo{Code: "AI_DISABLED", Message: "AI features are disabled"}

	AI_INPUT_TOO_LARGE = MessageInfo{Code: "AI_INPUT_TOO_LARGE", Message: "Input size exceeds the maximum allowed"}

	AI_PING_FAILED = MessageInfo{Code: "AI_PING_FAILED", Message: "AI ping failed"}	
)
