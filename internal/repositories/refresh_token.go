package repositories

import (
	"errors"
	"main/errs"
	"main/internal/models"
	"time"

	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token models.RefreshToken) (*models.RefreshToken, error)
	FindByToken(token string) (*models.RefreshToken, error)
	DeleteByToken(token string) error
	DeleteExpired() error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token models.RefreshToken) (*models.RefreshToken, error) {
	if err := r.db.Create(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) FindByToken(token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	if err := r.db.Preload("User").Where("token = ?", token).First(&rt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ERR_NOT_FOUND
		}
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.RefreshToken{}).Error
}

func (r *refreshTokenRepository) DeleteExpired() error {
	return r.db.Where("expired_at < ?", time.Now()).Delete(&models.RefreshToken{}).Error
}
