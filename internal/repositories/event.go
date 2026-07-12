package repositories

import (
	"context"
	"errors"
	"fmt"
	"main/internal/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ProcessResult string

const (
	Accepted  ProcessResult = "accepted"
	Rejected  ProcessResult = "rejected"
	Duplicate ProcessResult = "duplicate"
)

type ProcessResultDetail struct {
	Result  ProcessResult
	Reason  string
	OrderID int64
	Status  string
}

type OrderEventRepository interface {
	ProcessSingleEventTx(ctx context.Context, event models.OrderEvent) (ProcessResultDetail, error)
	ProcessBatchEventsTx(ctx context.Context, events []models.OrderEvent) ([]ProcessResultDetail, error)
	UpdateDriverNote(ctx context.Context, orderID int64, note string) error
	DriverUpdateStatus(ctx context.Context, orderID int64, newStatus models.OrderStatus, updatedBy string) error
}

type orderEventRepository struct {
	db *gorm.DB
}

func NewOrderEventRepository(db *gorm.DB) *orderEventRepository {
	return &orderEventRepository{
		db: db,
	}
}

func (r *orderEventRepository) ProcessSingleEventTx(ctx context.Context, event models.OrderEvent) (ProcessResultDetail, error) {
	var result ProcessResultDetail

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Raw("SELECT id, current_status FROM orders WHERE id = ? FOR UPDATE", event.OrderID).Scan(&order).Error; err != nil {
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

// ProcessBatchEventsTx processes multiple events in a single database transaction
// using bulk SQL operations to minimize network roundtrips.
// Events MUST be pre-sorted by event_at within each order_id group.
func (r *orderEventRepository) ProcessBatchEventsTx(ctx context.Context, events []models.OrderEvent) ([]ProcessResultDetail, error) {
	results := make([]ProcessResultDetail, len(events))

	if len(events) == 0 {
		return results, nil
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Collect unique order IDs
		orderIDSet := make(map[int64]bool)
		for _, e := range events {
			orderIDSet[e.OrderID] = true
		}
		orderIDs := make([]int64, 0, len(orderIDSet))
		for id := range orderIDSet {
			orderIDs = append(orderIDs, id)
		}

		// 2. Bulk SELECT with FOR UPDATE (one query for all orders)
		var orders []models.Order
		if err := tx.Raw("SELECT id, current_status FROM orders WHERE id IN ? FOR UPDATE", orderIDs).Scan(&orders).Error; err != nil {
			return err
		}

		// Build a map of order_id -> current_status (mutable, tracks in-flight changes)
		statusMap := make(map[int64]models.OrderStatus, len(orders))
		foundOrders := make(map[int64]bool, len(orders))
		for _, o := range orders {
			statusMap[o.ID] = o.CurrentStatus
			foundOrders[o.ID] = true
		}

		// 3. Validate each event in Go and classify results
		var acceptedEvents []models.OrderEvent
		finalStatuses := make(map[int64]models.OrderStatus) // tracks the last accepted status per order

		for i, event := range events {
			if !foundOrders[event.OrderID] {
				results[i] = ProcessResultDetail{
					Result:  Rejected,
					Reason:  "Order not found",
					OrderID: event.OrderID,
					Status:  string(event.NewStatus),
				}
				continue
			}

			prevStatus := statusMap[event.OrderID]
			nextStatus := event.NewStatus

			if nextStatus == prevStatus {
				results[i] = ProcessResultDetail{
					Result:  Duplicate,
					Reason:  fmt.Sprintf("Order is already in status '%s'", prevStatus),
					OrderID: event.OrderID,
					Status:  string(event.NewStatus),
				}
				continue
			}

			if !models.IsValidTransition(prevStatus, nextStatus) {
				results[i] = ProcessResultDetail{
					Result:  Rejected,
					Reason:  fmt.Sprintf("Invalid transition from '%s' to '%s'", prevStatus, nextStatus),
					OrderID: event.OrderID,
					Status:  string(event.NewStatus),
				}
				continue
			}

			// Accepted: update the in-flight status so subsequent events for the same order
			// are validated against the new status
			event.PreviousStatus = prevStatus
			acceptedEvents = append(acceptedEvents, event)
			statusMap[event.OrderID] = nextStatus
			finalStatuses[event.OrderID] = nextStatus

			results[i] = ProcessResultDetail{
				Result:  Accepted,
				OrderID: event.OrderID,
				Status:  string(event.NewStatus),
			}
		}

		if len(acceptedEvents) == 0 {
			return nil
		}

		// 4. Bulk UPDATE orders using a single CASE statement
		if len(finalStatuses) > 0 {
			var caseBuilder strings.Builder
			updateIDs := make([]interface{}, 0, len(finalStatuses))

			caseBuilder.WriteString("UPDATE orders SET current_status = CASE id ")
			for id, status := range finalStatuses {
				caseBuilder.WriteString(fmt.Sprintf("WHEN %d THEN '%s' ", id, status))
				updateIDs = append(updateIDs, id)
			}
			caseBuilder.WriteString("END, updated_at = NOW() WHERE id IN (")
			for i := range updateIDs {
				if i > 0 {
					caseBuilder.WriteString(",")
				}
				caseBuilder.WriteString(fmt.Sprintf("%v", updateIDs[i]))
			}
			caseBuilder.WriteString(")")

			if err := tx.Exec(caseBuilder.String()).Error; err != nil {
				return err
			}
		}

		// 5. Bulk INSERT all accepted events (one query)
		if err := tx.Create(&acceptedEvents).Error; err != nil {
			return err
		}

		return nil
	})

	return results, err
}

// UpdateDriverNote cập nhật driver_note vào event mới nhất của đơn hàng.
// Không tạo event mới, không thay đổi status.
func (r *orderEventRepository) UpdateDriverNote(ctx context.Context, orderID int64, note string) error {
	result := r.db.WithContext(ctx).Model(&models.OrderEvent{}).
		Where("order_id = ?", orderID).
		Order("event_at DESC").
		Limit(1).
		Update("driver_note", note)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no event found for order")
	}
	return nil
}

// DriverUpdateStatus cho phép driver chuyển trạng thái đơn, nhưng chỉ khi đơn đang ở "packed".
// Tạo một event mới trong transaction, đảm bảo tính toàn vẹn dữ liệu.
func (r *orderEventRepository) DriverUpdateStatus(ctx context.Context, orderID int64, newStatus models.OrderStatus, updatedBy string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Raw("SELECT id, current_status FROM orders WHERE id = ? FOR UPDATE", orderID).Scan(&order).Error; err != nil {
			return err
		}
		if order.ID == 0 {
			return errors.New("order not found")
		}

		// Guard: driver chỉ được update khi đơn đang ở "packed"
		if !models.IsDriverAllowedFromStatus(order.CurrentStatus) {
			return fmt.Errorf("driver cannot update order in '%s' status: order must be in 'packed' status", order.CurrentStatus)
		}

		// Guard: chỉ cho phép target status hợp lệ cho driver
		if !models.IsDriverAllowedStatus(newStatus) {
			return fmt.Errorf("driver is not allowed to set status to '%s'", newStatus)
		}

		// Validate transition theo business rules
		if !models.IsValidTransition(order.CurrentStatus, newStatus) {
			return fmt.Errorf("invalid transition from '%s' to '%s'", order.CurrentStatus, newStatus)
		}

		prevStatus := order.CurrentStatus
		now := time.Now()

		if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Update("current_status", newStatus).Error; err != nil {
			return err
		}

		event := models.OrderEvent{
			OrderID:        orderID,
			PreviousStatus: prevStatus,
			NewStatus:      newStatus,
			UpdatedBy:      updatedBy,
			EventAt:        now,
		}
		return tx.Create(&event).Error
	})
}
