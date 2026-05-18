package repositories

import (
	"errors"
	"fmt"
	"main/internal/models"

	"gorm.io/gorm"
)

type ProcessResult string

const (
	Accepted  ProcessResult = "accepted"
	Rejected  ProcessResult = "rejected"
	Duplicate ProcessResult = "duplicate"
)

type ProcessResultDetail struct {
	Result ProcessResult
	Reason string
}

type OrderEventRepository interface {
	ProcessSingleEventTx(event models.OrderEvent) (ProcessResultDetail, error)
}

type orderEventRepository struct {
	db *gorm.DB
}

func NewOrderEventRepository(db *gorm.DB) *orderEventRepository {
	return &orderEventRepository{
		db: db,
	}
}

func (r *orderEventRepository) ProcessSingleEventTx(event models.OrderEvent) (ProcessResultDetail, error) {
	var result ProcessResultDetail

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Raw("SELECT * FROM orders WHERE id = ? FOR UPDATE", event.OrderID).Scan(&order).Error; err != nil {
			return err
		}

		if order.ID == 0 {
			result = ProcessResultDetail{
				Result: Rejected,
				Reason: "Order not found",
			}
			return nil
		}

		prevStatus := order.CurrentStatus
		nextStatus := event.NewStatus

		if nextStatus == prevStatus {
			result = ProcessResultDetail{
				Result: Duplicate,
				Reason: fmt.Sprintf("Order is already in status '%s'", prevStatus),
			}
			return nil
		}

		if !models.IsValidTransition(prevStatus, nextStatus) {
			result = ProcessResultDetail{
				Result: Rejected,
				Reason: fmt.Sprintf("Invalid transition from '%s' to '%s'", prevStatus, nextStatus),
			}
			return nil
		}

		if err := tx.Model(&models.Order{}).Where("id = ?", event.OrderID).Update("current_status", nextStatus).Error; err != nil {
			return err
		}

		event.PreviousStatus = prevStatus
		if err := tx.Create(&event).Error; err != nil {
			return err
		}

		result = ProcessResultDetail{
			Result: Accepted,
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ProcessResultDetail{
				Result: Rejected,
				Reason: "Order not found",
			}, nil
		}
		return ProcessResultDetail{}, err
	}

	return result, nil
}
