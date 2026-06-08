package dto

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email" example:"admin@order.com"`
	Password string `json:"password" validate:"required"       example:"12345"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Role        string `json:"role"`
}
