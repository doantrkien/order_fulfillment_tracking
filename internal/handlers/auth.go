package handlers

import (
	"errors"
	"main/constant"
	"main/errs"

	dto "main/internal/dto/api"
	"main/internal/services"
	"main/response"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login godoc
// @Summary Login to get access token
// @Description Authenticate with email and password to receive a JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} response.ResponseStruct{data=dto.LoginResponse}
// @Failure 400 {object} response.ErrorBadReqResponse
// @Failure 401 {object} response.ErrorUnauthenticatedResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}
	if err := validator.New().Struct(req); err != nil {
		return response.ResponseError(c, errs.ERR_INVALID_INPUT, nil)
	}

	resp, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, errs.ERR_INVALID_CREDENTAIL) {
			return response.ResponseError(c, errs.ERR_INVALID_CREDENTAIL, nil)
		}
		return response.ResponseError(c, errs.ERR_INTERNAL_SERVER, nil)
	}
	return response.ResponseSuccess(c, 200, constant.SUCCESS.Message, resp)
}
