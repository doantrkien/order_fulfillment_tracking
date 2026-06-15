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

	GEMINI_API_KEY_EMPTY = MessageInfo{Code: "GEMINI_API_KEY_EMPTY", Message: "Gemini API key is empty"}

	GEMINI_CLIENT_CREATE_FAILED = MessageInfo{Code: "GEMINI_CLIENT_CREATE_FAILED", Message: "Failed to create gemini client"}

	GEMINI_CLIENT_TIMEOUT = MessageInfo{Code: "GEMINI_CLIENT_TIMEOUT", Message: "Gemini client timeout"}

	GEMINI_CLIENT_RETRY_LIMIT = MessageInfo{Code: "GEMINI_CLIENT_RETRY_LIMIT", Message: "Gemini client retry limit"}

	GEMINI_GENERATE_CONTENT_FAILED = MessageInfo{Code: "GEMINI_GENERATE_CONTENT_FAILED", Message: "Gemini generate content failed"}

	AI_RESPONSE_VALIDATION_FAILED = MessageInfo{Code: "AI_RESPONSE_VALIDATION_FAILED", Message: "AI response validation failed"}
	
	AI_DISABLED        = MessageInfo{Code: "AI_DISABLED", Message: "AI features are disabled"}

	AI_INPUT_TOO_LARGE = MessageInfo{Code: "AI_INPUT_TOO_LARGE", Message: "Input size exceeds the maximum allowed"}

	AI_PING_FAILED = MessageInfo{Code: "AI_PING_FAILED", Message: "AI ping failed"}	
)
