package dto

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email" example:"admin@demo.com"`
	Password string `json:"password" validate:"required"       example:"Admin@123"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"` // "Bearer"
	ExpiresIn   int    `json:"expires_in"` // seconds, e.g. 86400
	Role        string `json:"role"`       // "admin" or "driver"
}
