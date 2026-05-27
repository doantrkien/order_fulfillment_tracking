package response

// Error400Response represents a 400 Bad Request error
type Error400Response struct {
	Status  int    `json:"status" example:"400"`
	Message string `json:"message" example:"Invalid input data"`
	Data    any    `json:"data"`
}

// Error401Response represents a 401 Unauthorized error
type Error401Response struct {
	Status  int    `json:"status" example:"401"`
	Message string `json:"message" example:"Authentication required"`
	Data    any    `json:"data"`
}

// Error403Response represents a 403 Forbidden error
type Error403Response struct {
	Status  int    `json:"status" example:"403"`
	Message string `json:"message" example:"Permission denied"`
	Data    any    `json:"data"`
}

// Error404Response represents a 404 Not Found error
type Error404Response struct {
	Status  int    `json:"status" example:"404"`
	Message string `json:"message" example:"Resource not found"`
	Data    any    `json:"data"`
}

// Error500Response represents a 500 Internal Server Error
type Error500Response struct {
	Status  int    `json:"status" example:"500"`
	Message string `json:"message" example:"Internal server error"`
	Data    any    `json:"data"`
}
