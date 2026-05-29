package middleware

import (
	"main/errs"
	"main/internal/services"
	"main/response" // Sử dụng lại package response của bạn
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Authenticate: Kiểm tra JWT Token có hợp lệ không
func Authenticate() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Lấy token từ header "Authorization"
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.ResponseError(c, errs.ERR_UNAUTHORIZED, nil)
		}

		// Định dạng chuẩn: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return response.ResponseError(c, errs.ERR_UNAUTHORIZED, nil)
		}

		tokenString := parts[1]
		claims := &services.JWTClaims{}

		// Giải mã và verify chữ ký token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// Nếu lỗi hoặc token hết hạn/không hợp lệ
		if err != nil || !token.Valid {
			return response.ResponseError(c, errs.ERR_UNAUTHORIZED, nil)
		}

		// Lưu UserID (int64) và Role (string) vào Context để Handler sử dụng lại
		// Ví dụ hàm Logout của bạn đang gọi: c.Locals("user_id").(int64)
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)

		// Cho phép request đi tiếp vào Handler
		return c.Next()
	}
}

// Authorize: Phân quyền truy cập dựa trên Role
func Authorize(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Lấy Role từ Context (được gán ở Authenticate middleware)
		userRole := c.Locals("role")
		if userRole == nil {
			return response.ResponseError(c, errs.ERR_UNAUTHORIZED, nil)
		}

		roleStr := userRole.(string)

		// Kiểm tra xem user có nằm trong danh sách quyền được phép không
		for _, role := range allowedRoles {
			if roleStr == role {
				return c.Next() // Hợp lệ, đi tiếp
			}
		}

		// Nếu không khớp quyền nào (Ví dụ: customer muốn vào route của admin)
		// Trả về lỗi 403 Forbidden (Giả sử bạn có mã lỗi ERR_FORBIDDEN trong errs, nếu không có thể đổi thành ERR_UNAUTHORIZED)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "Bạn không có quyền truy cập vào chức năng này",
		})
	}
}
