package repositories

import (
	"main/internal/models"

	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (r *OrderRepository) GetAllOrder() ([]models.Order, error) {
	var orders []models.Order

	if err := r.db.Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) GetOrderDetail(id int) (*models.Order, error) {
	var order models.Order

	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

// func CreateOrder(order models.Order) (*models.Order, error) {
// 	return nil, nil
// }

// func UpdateOrderStatus(id uint, status string) (*models.Order, error) {
// 	return nil, nil
// }
