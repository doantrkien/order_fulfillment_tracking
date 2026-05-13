package repositories

import (
	"errors"
	"main/internal/dto"
	"main/internal/models"
	"main/pkg/utils/constant"

	"gorm.io/gorm"
)

type OrderRepository interface {
	GetAllOrder(query dto.OrderQuery) ([]models.Order, int64, error)
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

func (r *orderRepository) GetAllOrder(query dto.OrderQuery) ([]models.Order, int64, error) {

	var (
		orders []models.Order
		total  int64
	)

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}

	offset := (query.Page - 1) * query.Limit

	db := r.db.Model(&models.Order{})

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if query.Date != "" {
		db = db.Where("DATE(created_at) = ?", query.Date)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.
		Limit(query.Limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
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
