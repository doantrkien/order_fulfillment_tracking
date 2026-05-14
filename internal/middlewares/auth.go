package middlewares

import (
	"fmt"
	"main/pkg/utils/constant"
	"main/pkg/utils/response"
	"os"

	"github.com/gofiber/fiber/v3"
)

var routePermissions = map[string]string{
	"POST:/api/v1/orders":             "order.create",
	"GET:/api/v1/orders":              "order.read",
	"GET:/api/v1/orders/:id":          "order.read",
	"PATCH:/api/v1/orders/:id/status": "order.update",

	"POST:/api/v1/order-events/import": "event.create",

	"GET:/api/v1/reports/daily": "report.read",
}

var rolePermissions = map[string][]string{

	"admin": {
		"order.create",
		"order.read",
		"order.update",
		"event.create",
		"report.read",
	},

	"shipper": {
		"order.read",
		"order.update",
	},

	"customer": {
		"order.create",
		"order.read",
	},

	"operator": {
		"event.create",
		"report.read",
		"order.update",
	},
}

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

func Authorize() fiber.Handler {
	return func(c fiber.Ctx) error {

		role, ok := c.Locals("role").(string)

		if !ok {
			return response.Reponse(c, 401, constant.UN_AUTHENTICATION, nil)
		}

		currentRoute := c.Method() + ":" + c.Path()

		fmt.Println(currentRoute)

		requiredPermission, ok := routePermissions[currentRoute]

		if !ok {
			return response.Reponse(c, 404, constant.URL_NOT_FOUND, nil)
		}

		permissions := rolePermissions[role]

		for _, permission := range permissions {

			if permission == requiredPermission {
				return c.Next()
			}
		}

		return response.Reponse(
			c,
			403,
			constant.UN_AUTHORIZATION,
			nil,
		)
	}
}
