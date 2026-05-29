package handlers

import (
	"errors"
	"main/constant"
	"main/errs"
	"main/internal/dto"
	"main/internal/services"
	"main/response"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login godoc
// @Summary User login
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login credentials"
// @Success 200 {object} response.ResponseStruct{data=dto.LoginResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest

	// Chuẩn Fiber v3: Dùng c.Bind().JSON thay cho c.BodyParser
	if err := c.Bind().JSON(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	// Sửa: Truyền nguyên struct `req` thay vì truyền rời username, password
	loginRes, err := h.authService.Login(req)
	if err != nil {
		if errors.Is(err, errs.ERR_UNAUTHENTICATED) || errors.Is(err, errs.ERR_UNAUTHORIZED) {
			return response.ResponseError(c, errs.ERR_UNAUTHORIZED, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	// Sửa: loginRes đã là kiểu dto.LoginResponse rồi, truyền thẳng vào luôn
	return response.ResponseSuccess(c, http.StatusOK, constant.SUCCESS.Message, loginRes)
}

// Logout godoc
// @Summary User logout
// @Tags Auth
// @Accept json
// @Produce json
// @Param logout body dto.LogoutRequest true "Logout credentials"
// @Success 200 {object} response.ResponseStruct
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req dto.LogoutRequest

	// Sửa: Đọc RefreshToken từ Body thông qua dto.LogoutRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	// Sửa: Truyền chuỗi RefreshToken vào hàm Logout theo đúng định nghĩa Service
	err := h.authService.Logout(req.RefreshToken)
	if err != nil {
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	return response.ResponseSuccess(c, http.StatusOK, constant.SUCCESS.Message, nil)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.ResponseStruct{data=dto.LoginResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Failure 403 {object} response.ErrorUnauthorizedResponse
// @Failure 500 {object} response.ErrorInternalServerErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	var req dto.RefreshTokenRequest

	// Chuẩn Fiber v3: Dùng c.Bind().JSON thay cho c.BodyParser
	if err := c.Bind().JSON(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	refreshRes, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, errs.ERR_UNAUTHENTICATED) || errors.Is(err, errs.ERR_UNAUTHORIZED) {
			return response.ResponseError(c, errs.ERR_UNAUTHORIZED, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}

	// Sửa: refreshRes đã là kiểu dto.LoginResponse rồi, truyền thẳng vào luôn
	return response.ResponseSuccess(c, http.StatusOK, constant.SUCCESS.Message, refreshRes)
}
