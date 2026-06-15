package response

import (
	"errors"
	"main/errs"

	"github.com/gofiber/fiber/v3"
)

type ResponseStruct struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func ResponseSuccess(c fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(ResponseStruct{
		Status:  "SUCCESS",
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
	Status     string     `json:"status" example:"SUCCESS"`
	Message    string     `json:"message" example:"Success"`
	Pagination Pagination `json:"pagination"`
	Data       any        `json:"data"`
}

func PaginatedSuccess(c fiber.Ctx, message string, data any, currentPage, limitItems int, totalItems int64) error {
	return c.Status(200).JSON(PaginatedResponse{
		Status:  "SUCCESS",
		Message: message,
		Pagination: Pagination{
			CurrentPage: currentPage,
			LimitItems:  limitItems,
			TotalItems:  totalItems,
		},
		Data: data,
	})
}

func ResponseError(c fiber.Ctx, err error, data interface{}) error {
	var appErr *errs.AppError

	if errors.As(err, &appErr) {
		return c.Status(appErr.StatusCode).JSON(ResponseStruct{
			Status:  "ERROR",
			Data:    data,
			Message: appErr.Err.Message,
		})
	}

	return c.Status(500).JSON(ResponseStruct{
		Status:  "ERROR",
		Data:    data,
		Message: errs.ERR_INTERNAL_SERVER.Err.Message,
	})
}
