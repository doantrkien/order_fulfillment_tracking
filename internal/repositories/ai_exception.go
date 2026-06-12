package repositories

import (
	"context"
	"main/internal/models"

	"gorm.io/gorm"
)

type AIRepository interface {
	Save(ctx context.Context, aiException *models.AIException) error
	GetLatestByOrderID(ctx context.Context, orderID int64) (*models.AIException, error)
	GetAIContextByOrderID(ctx context.Context, orderID int64) (*models.AIContext, error)
}

type aiRepository struct {
	db *gorm.DB
}

func NewAIRepository(db *gorm.DB) AIRepository {
	return &aiRepository{db: db}
}

func (r *aiRepository) Save(ctx context.Context, aiException *models.AIException) error {
	return r.db.WithContext(ctx).Create(aiException).Error
}

func (r *aiRepository) GetLatestByOrderID(ctx context.Context, orderID int64) (*models.AIException, error) {
	var result models.AIException
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("evaluated_at DESC").
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAIContextByOrderID aggregates order + events into AIContext used for prompt building.
func (r *aiRepository) GetAIContextByOrderID(ctx context.Context, orderID int64) (*models.AIContext, error) {
	var order models.Order
	if err := r.db.WithContext(ctx).Where("id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}

	var events []models.OrderEvent
	if err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("event_at DESC").
		Find(&events).Error; err != nil {
		return nil, err
	}

	aiEvents := make([]models.AIEvent, 0, len(events))
	for _, e := range events {
		aiEvents = append(aiEvents, models.AIEvent{
			EventAt:        e.EventAt,
			PreviousStatus: e.PreviousStatus,
			NewStatus:      e.NewStatus,
			DriverID:       e.DriverID,
			DriverNote:     e.DriverNote,
			UpdatedBy:      e.UpdatedBy,
		})
	}

	return &models.AIContext{
		OrderID:       order.ID,
		CreatedAt:     order.CreatedAt,
		CurrentStatus: order.CurrentStatus,
		TotalAmount:   order.TotalAmount,
		PaymentStatus: models.DerivePaymentStatus(order.CurrentStatus),
		RefundStatus:  models.DeriveRefundStatus(order.CurrentStatus),
		Events:        aiEvents,
	}, nil
}
