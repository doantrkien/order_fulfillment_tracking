package middlewares

import (
	"main/errs"
	"main/response"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func Authenticate() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return response.ResponseError(c, errs.ERR_UNAUTHENTICATED, nil)
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "default-secret-key-change-in-production"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return response.ResponseError(c, errs.ERR_UNAUTHENTICATED, nil)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return response.ResponseError(c, errs.ERR_UNAUTHENTICATED, nil)
		}

		role, _ := claims["role"].(string)
		sub, _ := claims["sub"].(float64)

		c.Locals("role", role)
		c.Locals("user_id", int64(sub))

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
