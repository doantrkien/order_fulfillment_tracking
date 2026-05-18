package repositories

import (
	"errors"
	"main/internal/models"
	"main/pkg/utils/constant"

	"gorm.io/gorm"
)

type OrderRepository interface {
	GetAllOrder() ([]models.Order, error)
	GetOrderDetail(id int) (*models.Order, error)
	CreateOrder(order models.Order) (*models.Order, error)
	UpdateOrderStatus(id int64, status string) (*models.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *orderRepository {
	return &orderRepository{
		db: db,
	}
}

func (r *orderRepository) GetAllOrder() ([]models.Order, error) {
	var orders []models.Order

	if err := r.db.Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepository) GetOrderDetail(id int) (*models.Order, error) {
	var order models.Order

	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *orderRepository) CreateOrder(order models.Order) (*models.Order, error) {
	if err := r.db.Create(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) UpdateOrderStatus(id int64, status string) (*models.Order, error) {
	var order models.Order

	if err := r.db.First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constant.ERR_NOT_FOUND
		}
		return nil, err
	}

	order.Status = models.OrderStatus(status)

	if err := r.db.Save(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}
