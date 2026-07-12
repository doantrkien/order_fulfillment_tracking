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
	GetAllOrder(query dto.OrderQuery) ([]models.Order, int64, error)
	GetOrderDetail(int64) (*models.Order, error)
	IsDriverAssignedToOrder(orderID int64, driverID int64) (bool, error)
	CreateOrder(order models.Order, updatedBy string) (*models.Order, error)
	UpdateOrderStatus(id int64, status, updatedBy string, driverID *int64) (*models.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *orderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) GetAllOrder(query dto.OrderQuery) ([]models.Order, int64, error) {
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

	if query.DriverID != 0 {
		db = db.Where("EXISTS (SELECT 1 FROM order_events WHERE order_events.order_id = orders.id AND order_events.driver_id = ?)", query.DriverID)
	}
	if query.Status != "" {
		db = db.Where("orders.current_status = ?", query.Status)
	}
	if query.Date != "" {
		db = db.Where("DATE(orders.created_at) = ?", query.Date)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Limit(query.LimitItems).Offset(offset).Order("orders.created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepository) GetOrderDetail(id int64) (*models.Order, error) {

	var order models.Order

	if err := r.db.First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ERR_NOT_FOUND
		}

		return nil, err
	}

	// Fetch the driver_note from the latest event (if any)
	var latestNote struct {
		DriverNote *string
	}
	r.db.Raw(`SELECT driver_note FROM order_events WHERE order_id = ? ORDER BY event_at DESC LIMIT 1`, id).Scan(&latestNote)
	order.LatestDriverNote = latestNote.DriverNote

	return &order, nil
}

func (r *orderRepository) IsDriverAssignedToOrder(orderID int64, driverID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.OrderEvent{}).
		Where("order_id = ? AND driver_id = ?", orderID, driverID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *orderRepository) CreateOrder(order models.Order, updatedBy string) (*models.Order, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		event := models.OrderEvent{
			OrderID:   order.ID,
			NewStatus: order.CurrentStatus,
			UpdatedBy: updatedBy,
			EventAt:   time.Now(),
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

func (r *orderRepository) UpdateOrderStatus(id int64, status, updatedBy string, driverID *int64) (*models.Order, error) {

	var updatedOrder models.Order
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var order models.Order
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
			DriverID:       driverID,
		}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}

		updatedOrder = order
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &updatedOrder, nil
}
