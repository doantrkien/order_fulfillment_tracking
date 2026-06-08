package repositories

import (
	"errors"
	"main/errs"
	"main/internal/models"

	"gorm.io/gorm"
)

type AccountRepository interface {
	FindByEmail(email string) (*models.Account, error)
	Create(account models.Account) (*models.Account, error)
}

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *accountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) FindByEmail(email string) (*models.Account, error) {
	var account models.Account

	if err := r.db.Where("email = ?", email).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ERR_NOT_FOUND
		}
		return nil, err
	}

	return &account, nil
}

func (r *accountRepository) Create(account models.Account) (*models.Account, error) {
	if err := r.db.Create(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}
