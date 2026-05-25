package response

import (
	"errors"
	"main/errs"

	"github.com/gofiber/fiber/v3"
)

type ResponseStruct struct {
	Status  int         `json:"status" example:"200"`
	Data    interface{} `json:"data"`
	Message string      `json:"message" example:"Success"`
}

func ResponseSuccess(c fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(ResponseStruct{
		Status:  status,
		Message: message,
		Data:    data,
	})
}

type Pagination struct {
	CurrentPage int   `json:"current_page" example:"1"`
	LimitItems  int   `json:"limit_items" example:"10"`
	TotalItems  int64 `json:"total_items" example:"100"`
}

type PaginatedResponse struct {
	Status     int        `json:"status" example:"200"`
	Message    string     `json:"message" example:"Success"`
	Pagination Pagination `json:"pagination"`
	Data       any        `json:"data"`
}

func PaginatedSuccess(c fiber.Ctx, message string, data any, currentPage, limitItems int, totalItems int64) error {
	return c.Status(200).JSON(PaginatedResponse{
		Status:  200,
		Message: message,
		Pagination: Pagination{
			CurrentPage: currentPage,
			LimitItems:  limitItems,
			TotalItems:  totalItems,
		},
		Data: data,
	})
}

func ResponseError(c fiber.Ctx, err error) error {
	var appErr *errs.AppError

	if errors.As(err, &appErr) {
		return c.Status(appErr.StatusCode).JSON(ResponseStruct{
			Status:  appErr.StatusCode,
			Data:    nil,
			Message: appErr.Err.Message,
		})
	}

	return c.Status(500).JSON(ResponseStruct{
		Status:  500,
		Data:    nil,
		Message: errs.ERR_INTERNAL_SERVER.Err.Message,
	})
}
