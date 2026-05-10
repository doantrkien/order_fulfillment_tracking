package repositories

import (
	"main/internal/dto"
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

func (r *OrderRepository) CreateOrder(req dto.OrderRequest) (*models.Order, error) {
	order := models.Order{}

	if err := r.db.Create(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) UpdateOrderStatus(id uint, status string) (*models.Order, error) {
	var order models.Order

	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}

	// order.Status = status

	if err := r.db.Save(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}
