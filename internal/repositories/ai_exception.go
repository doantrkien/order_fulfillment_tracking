package repositories

import (
	"context"
	"encoding/json"
	"main/internal/models"
	"strings"

	"gorm.io/gorm"
)

type AIRepository interface {
	Save(ctx context.Context, aiException *models.AIException) error
	GetLatestAnalysisByOrderID(ctx context.Context, orderID int64) (*models.AIException, error)
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

func (r *aiRepository) GetLatestAnalysisByOrderID(ctx context.Context, orderID int64) (*models.AIException, error) {
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

func (r *aiRepository) GetAIContextByOrderID(ctx context.Context, orderID int64) (*models.AIContext, error) {

	var order models.Order
	if err := r.db.WithContext(ctx).Where("id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}

	var events []models.OrderEvent
	if err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("event_at ASC").
		Find(&events).Error; err != nil {
		return nil, err
	}

	aiEvents := make([]models.AIEvent, 0, len(events))
	var driverNotesParts []string
	for _, e := range events {
		aiEvents = append(aiEvents, models.AIEvent{
			EventAt:        e.EventAt,
			PreviousStatus: e.PreviousStatus,
			NewStatus:      e.NewStatus,
			DriverID:       e.DriverID,
			DriverNote:     e.DriverNote,
			UpdatedBy:      e.UpdatedBy,
		})
		// Aggregate all non-empty driver notes for rule-based analysis
		if e.DriverNote != nil && strings.TrimSpace(*e.DriverNote) != "" {
			driverNotesParts = append(driverNotesParts, strings.TrimSpace(*e.DriverNote))
		}
	}

	var userInfo models.UserInfo
	_ = json.Unmarshal(order.UserInfo, &userInfo)

	return &models.AIContext{
		OrderID:         order.ID,
		CreatedAt:       order.CreatedAt,
		CurrentStatus:   order.CurrentStatus,
		TotalAmount:     order.TotalAmount,
		CustomerName:    userInfo.Username,
		ShippingAddress: userInfo.ShippingAddress,
		PaymentStatus:   models.DerivePaymentStatus(order.CurrentStatus),
		RefundStatus:    models.DeriveRefundStatus(order.CurrentStatus),
		DriverNotes:     strings.Join(driverNotesParts, "; "),
		Events:          aiEvents,
	}, nil
}
