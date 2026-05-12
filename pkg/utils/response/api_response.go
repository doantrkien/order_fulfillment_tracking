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
