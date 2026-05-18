package middlewares

import (
	"main/pkg/utils/constant"
	"main/pkg/utils/response"
	"os"

	"github.com/gofiber/fiber/v3"
)

func Authenticate() fiber.Handler {
	return func(c fiber.Ctx) error {

		apiKey := c.Get("X-API-Key")

		if apiKey == "" {
			return response.Reponse(c, 401, constant.UN_AUTHENTICATION, nil)
		}

		switch apiKey {
		case os.Getenv("CUSTOMER_API_KEY"):
			c.Locals("role", "customer")

		case os.Getenv("DRIVER_API_KEY"):
			c.Locals("role", "shipper")

		case os.Getenv("ADMIN_API_KEY"):
			c.Locals("role", "admin")

		default:
			return response.Reponse(c, 401, constant.UN_AUTHENTICATION, nil)
		}

		return c.Next()
	}
}

func Authorize(roles []string) fiber.Handler {
	return func(c fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok {
			return response.Reponse(c, 401, constant.UN_AUTHENTICATION, nil)
		}

		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		return response.Reponse(c, 403, constant.UN_AUTHORIZATION, nil)
	}
}
