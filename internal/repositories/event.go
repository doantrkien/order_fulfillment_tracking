package repositories

import (
	"main/internal/models"

	"gorm.io/gorm"
)

type OrderEventRepository interface {
	ImportOrderEvents(events []models.OrderEvent) error
}

type orderEventRepository struct {
	db *gorm.DB
}

func NewOrderEventRepository(db *gorm.DB) *orderEventRepository {
	return &orderEventRepository{
		db: db,
	}
}

func (r *orderEventRepository) ImportOrderEvents(events []models.OrderEvent) error {
	// TODO
	return nil
}
