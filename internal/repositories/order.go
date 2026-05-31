package repositories

import (
	"errors"
	"main/errs"
	"main/internal/dto"
	"main/internal/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository interface {
	GetAllOrder(query dto.OrderQuery, role string, userID int64) ([]models.Order, int64, error)
	GetOrderDetail(id int64, role string, userID int64) (*models.Order, error)
	CreateOrder(order models.Order) (*models.Order, error)
	UpdateOrderStatus(id int64, status string, updatedBy string) (*models.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *orderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) GetAllOrder(query dto.OrderQuery, role string, userID int64) ([]models.Order, int64, error) {
	var (
		orders []models.Order
		total  int64
	)

	if query.PageNumber <= 0 {
		query.PageNumber = 1
	}
	if query.LimitItems <= 0 {
		query.LimitItems = 10
	}

	offset := (query.PageNumber - 1) * query.LimitItems
	db := r.db.Model(&models.Order{})

	if query.Status != "" {
		db = db.Where("current_status = ?", query.Status)
	}
	if query.Date != "" {
		db = db.Where("DATE(created_at) = ?", query.Date)
	}
	if role == "driver" {
		subQuery := r.db.Table("order_events oe").Select("oe.order_id").
			Joins(`JOIN (
				SELECT order_id, MAX(event_at) AS max_event_at
				FROM order_events
				GROUP BY order_id
			) latest ON latest.order_id = oe.order_id AND latest.max_event_at = oe.event_at`).
			Where("oe.driver_id = ?", userID)
		db = db.Where("id IN (?)", subQuery)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Limit(query.LimitItems).Offset(offset).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepository) GetOrderDetail(id int64, role string, userID int64) (*models.Order, error) {
	var order models.Order

	db := r.db.Model(&models.Order{}).Where("id = ?", id)
	if role == "driver" {
		subQuery := r.db.Table("order_events oe").Select("oe.order_id").
			Joins(`JOIN (
				SELECT order_id, MAX(event_at) AS max_event_at
				FROM order_events
				GROUP BY order_id
			) latest ON latest.order_id = oe.order_id AND latest.max_event_at = oe.event_at`).
			Where("oe.driver_id = ?", userID)
		db = db.Where("id IN (?)", subQuery)
	}

	if err := db.First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if role == "driver" {
				var existingOrder models.Order
				if err2 := r.db.Model(&models.Order{}).Where("id = ?", id).First(&existingOrder).Error; err2 == nil {
					return nil, errs.ERR_UNAUTHORIZED
				}
			}
			return nil, errs.ERR_NOT_FOUND
		}

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

// func (r *orderRepository) UpdateOrderStatus(id int64, status string) (*models.Order, error) {

// 	var order models.Order
// 	if err := r.db.First(&order, id).Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, errs.ERR_NOT_FOUND
// 		}
// 		return nil, err
// 	}

// 	order.CurrentStatus = models.OrderStatus(status)
// 	if err := r.db.Save(&order).Error; err != nil {
// 		return nil, err
// 	}
// 	return &order, nil
// }

func (r *orderRepository) UpdateOrderStatus(
	id int64,
	status string,
	updatedBy string,
) (*models.Order, error) {

	var order models.Order
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errs.ERR_NOT_FOUND
			}
			return err
		}

		if !models.IsValidTransition(order.CurrentStatus, models.OrderStatus(status)) {
			return errs.ERR_INVALID_STATUS
		}

		previousStatus := order.CurrentStatus

		order.CurrentStatus = models.OrderStatus(status)

		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		event := models.OrderEvent{
			OrderID:        order.ID,
			PreviousStatus: previousStatus,
			NewStatus:      order.CurrentStatus,
			UpdatedBy:      updatedBy,
			EventAt:        time.Now(),
		}

		if err := tx.Create(&event).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &order, nil
}
