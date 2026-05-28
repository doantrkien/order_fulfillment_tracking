package middlewares

import (
	"fmt"
	"main/errs"
	"main/response"
	"os"

	"github.com/gofiber/fiber/v3"
)

func Authenticate() fiber.Handler {
	return func(c fiber.Ctx) error {

		apiKey := c.Get("X-API-KEY")
		fmt.Printf("Received API Key: %s\n", apiKey) // Debug log for received API key

		if apiKey == "" {
			return response.ResponseError(c, errs.ERR_UNAUTHENTICATED, nil)
		}

		switch apiKey {
		case os.Getenv("CUSTOMER_API_KEY"):
			c.Locals("role", "customer")

		case os.Getenv("DRIVER_API_KEY"):
			c.Locals("role", "driver")

		case os.Getenv("ADMIN_API_KEY"):
			c.Locals("role", "admin")

		default:
			return response.ResponseError(c, errs.ERR_UNAUTHENTICATED, nil)
		}

		return c.Next()
	}
}

func Authorize(roles []string) fiber.Handler {
	return func(c fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok {
			return response.ResponseError(c, errs.ERR_UNAUTHENTICATED, nil)
		}

		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		return response.ResponseError(c, errs.ERR_UNAUTHORIZED, nil)
	}
}
