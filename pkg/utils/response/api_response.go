package response

import "github.com/gofiber/fiber/v3"

type ResponseStruct struct {
	Status  int         `json:"status"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func Reponse(c fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(ResponseStruct{
		Status:  status,
		Message: message,
		Data:    data,
	})
}

type Pagination struct {
	Page       int   `json:"current_page"`
	Limit      int   `json:"limit_item"`
	Total      int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse struct {
	Status     int        `json:"status"`
	Message    string     `json:"message"`
	Pagination Pagination `json:"pagination"`
	Data       any        `json:"data"`
}

func PaginatedSuccess(c fiber.Ctx, message string, data any, page, limit int, total int64) error {

	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return c.Status(200).JSON(PaginatedResponse{
		Status:  200,
		Message: message,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
		Data: data,
	})
}
