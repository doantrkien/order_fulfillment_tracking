package repositories

import (
	"context"
	"main/internal/models"

	"gorm.io/gorm"
)

type AIDraftRepository interface {
	Save(ctx context.Context, draft *models.AICustomerUpdateDraft) error
	GetLatestByOrderID(ctx context.Context, orderID int64) (*models.AICustomerUpdateDraft, error)
	GetPendingByOrderID(ctx context.Context, orderID int64) ([]models.AICustomerUpdateDraft, error)
	UpdateReviewStatus(ctx context.Context, id int64, status string, reviewedBy string) error
}

type aiDraftRepository struct {
	db *gorm.DB
}

func NewAIDraftRepository(db *gorm.DB) AIDraftRepository {
	return &aiDraftRepository{db: db}
}

func (r *aiDraftRepository) Save(ctx context.Context, draft *models.AICustomerUpdateDraft) error {
	return r.db.WithContext(ctx).Create(draft).Error
}

func (r *aiDraftRepository) GetLatestByOrderID(ctx context.Context, orderID int64) (*models.AICustomerUpdateDraft, error) {
	var draft models.AICustomerUpdateDraft
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		First(&draft).Error
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *aiDraftRepository) GetPendingByOrderID(ctx context.Context, orderID int64) ([]models.AICustomerUpdateDraft, error) {
	var drafts []models.AICustomerUpdateDraft
	err := r.db.WithContext(ctx).
		Where("order_id = ? AND review_status = ?", orderID, models.DraftReviewStatusPending).
		Order("created_at DESC").
		Find(&drafts).Error
	if err != nil {
		return nil, err
	}
	return drafts, nil
}

func (r *aiDraftRepository) UpdateReviewStatus(ctx context.Context, id int64, status string, reviewedBy string) error {
	return r.db.WithContext(ctx).
		Model(&models.AICustomerUpdateDraft{}).
		Where("id = ? AND review_status = ?", id, models.DraftReviewStatusPending).
		Updates(map[string]interface{}{
			"review_status": status,
			"reviewed_by":   reviewedBy,
			"reviewed_at":   gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}
