package response

// represents a 400 Bad Request error
type ErrorBadReqResponse struct {
	Status  int         `json:"status" example:"400"`
	Message string      `json:"message" example:"Invalid input data"`
	Data    interface{} `json:"data,omitempty" swaggertype:"object"`
}

// represents a 401 Unauthorized error
type ErrorUnauthenticatedResponse struct {
	Status  int         `json:"status" example:"401"`
	Message string      `json:"message" example:"Authentication required"`
	Data    interface{} `json:"data,omitempty" swaggertype:"object"`
}

// represents a 403 Forbidden error
type ErrorUnauthorizedResponse struct {
	Status  int         `json:"status" example:"403"`
	Message string      `json:"message" example:"Permission denied"`
	Data    interface{} `json:"data,omitempty" swaggertype:"object"`
}

// represents a 404 Not Found error
type ErrorNotFoundResponse struct {
	Status  int         `json:"status" example:"404"`
	Message string      `json:"message" example:"Resource not found"`
	Data    interface{} `json:"data,omitempty" swaggertype:"object"`
}

// represents a 500 Internal Server Error
type ErrorInternalServerErrorResponse struct {
	Status  int         `json:"status" example:"500"`
	Message string      `json:"message" example:"Internal server error"`
	Data    interface{} `json:"data,omitempty" swaggertype:"object"`
}
