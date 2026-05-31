package services

import (
	"errors"
	"fmt"
	"main/errs"
	"main/internal/dto"
	"main/internal/repositories"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(email, password string) (*dto.LoginResponse, error)
}

type authService struct {
	userRepo repositories.AccountRepository
}

func NewAuthService(userRepo repositories.AccountRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Login(email, password string) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, errs.ERR_NOT_FOUND) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ERR_INVALID_CREDENTAIL
		}
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("12345"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(hash))

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errs.ERR_INVALID_CREDENTAIL
	}

	expireHours := 24
	if val := os.Getenv("JWT_EXPIRE_HOURS"); val != "" {
		if h, err := strconv.Atoi(val); err == nil && h > 0 {
			expireHours = h
		}
	}

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Duration(expireHours) * time.Hour).Unix(),
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key-change-in-production"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken: signedToken,
		TokenType:   "Bearer",
		ExpiresIn:   expireHours * 3600,
		Role:        user.Role,
	}, nil
}
