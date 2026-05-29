package services

import (
	"main/errs"
	"main/internal/dto"
	"main/internal/models"
	"main/internal/repositories"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(req dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(refreshToken string) (*dto.LoginResponse, error)
	Logout(refreshToken string) error
}

type authService struct {
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
}

func NewAuthService(
	userRepo repositories.UserRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func getAccessExpire() time.Duration {
	val := os.Getenv("JWT_ACCESS_EXPIRE")
	if val == "" {
		return 15 * time.Minute
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}

func getRefreshExpire() time.Duration {
	val := os.Getenv("JWT_REFRESH_EXPIRE")
	if val == "" {
		return 7 * 24 * time.Hour
	}
	// support days: "7d"
	if len(val) > 1 && val[len(val)-1] == 'd' {
		days, err := strconv.Atoi(val[:len(val)-1])
		if err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 7 * 24 * time.Hour
	}
	return d
}

func generateAccessToken(user *models.User) (string, time.Duration, error) {
	expire := getAccessExpire()
	claims := JWTClaims{
		UserID: user.ID,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", 0, err
	}
	return signed, expire, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errs.ERR_UNAUTHENTICATED
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errs.ERR_UNAUTHENTICATED
	}

	accessToken, expire, err := generateAccessToken(user)
	if err != nil {
		return nil, errs.ERR_INTERNAL_SERVER
	}

	refreshExpire := getRefreshExpire()
	rt := models.RefreshToken{
		UserID:    user.ID,
		Token:     uuid.NewString(),
		ExpiredAt: time.Now().Add(refreshExpire),
	}
	created, err := s.refreshTokenRepo.Create(rt)
	if err != nil {
		return nil, errs.ERR_INTERNAL_SERVER
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: created.Token,
		ExpiresIn:    int64(expire.Seconds()),
	}, nil
}

func (s *authService) RefreshToken(refreshToken string) (*dto.LoginResponse, error) {
	rt, err := s.refreshTokenRepo.FindByToken(refreshToken)
	if err != nil {
		return nil, errs.ERR_UNAUTHENTICATED
	}

	if time.Now().After(rt.ExpiredAt) {
		s.refreshTokenRepo.DeleteByToken(refreshToken)
		return nil, errs.ERR_UNAUTHENTICATED
	}

	accessToken, expire, err := generateAccessToken(rt.User)
	if err != nil {
		return nil, errs.ERR_INTERNAL_SERVER
	}

	// Rotate refresh token
	s.refreshTokenRepo.DeleteByToken(refreshToken)
	newRT := models.RefreshToken{
		UserID:    rt.UserID,
		Token:     uuid.NewString(),
		ExpiredAt: rt.ExpiredAt, // giữ nguyên thời hạn gốc
	}
	created, err := s.refreshTokenRepo.Create(newRT)
	if err != nil {
		return nil, errs.ERR_INTERNAL_SERVER
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: created.Token,
		ExpiresIn:    int64(expire.Seconds()),
	}, nil
}

func (s *authService) Logout(refreshToken string) error {
	return s.refreshTokenRepo.DeleteByToken(refreshToken)
}
