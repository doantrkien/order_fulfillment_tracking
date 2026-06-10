package repositories

import (
	"context"
	"main/internal/models" // Đường dẫn của bạn

	"gorm.io/gorm"
)

// 1. Định nghĩa Interface (Thêm context và hàm Get)
type AIRepository interface {
	Save(ctx context.Context, aiException *models.AIException) error
	GetLatestByOrderID(ctx context.Context, orderID int64) (*models.AIException, error)
	GetAIContextByOrderID(ctx context.Context, orderID int64) (*models.AIContext, error)
}

type aiRepository struct {
	db *gorm.DB
}

// 2. Constructor trả về Interface thay vì struct
func NewAIRepository(db *gorm.DB) AIRepository {
	return &aiRepository{
		db: db,
	}
}

// 3. Hàm Save có sử dụng context
func (r *aiRepository) Save(ctx context.Context, aiException *models.AIException) error {
	// Thêm .WithContext(ctx) vào trước lệnh gọi DB
	if err := r.db.WithContext(ctx).Create(aiException).Error; err != nil {
		return err
	}
	return nil
}

// 4. Bổ sung hàm lấy kết quả AI mới nhất của một đơn hàng
func (r *aiRepository) GetLatestByOrderID(ctx context.Context, orderID int64) (*models.AIException, error) {
	var result models.AIException

	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("evaluated_at DESC"). // Lấy cái mới nhất
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
		Order("event_at ASC").
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

	ctxObj := &models.AIContext{
		OrderID:       order.ID,
		CreatedAt:     order.CreatedAt,
		CurrentStatus: order.CurrentStatus,
		TotalAmount:   order.TotalAmount,
		Events:        aiEvents,
	}

	return ctxObj, nil
}
