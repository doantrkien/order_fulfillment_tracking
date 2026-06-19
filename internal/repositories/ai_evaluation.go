package repositories

import (
	"context"
	"main/internal/models"

	"gorm.io/gorm"
)

// AIEvaluationRepository handles persistence for evaluation batch runs and details.
type AIEvaluationRepository interface {
	CreateRun(ctx context.Context, run *models.AIEvaluationRun) error
	UpdateRun(ctx context.Context, run *models.AIEvaluationRun) error
	SaveDetail(ctx context.Context, detail *models.AIEvaluationResultDetail) error
	GetRunByID(ctx context.Context, id int64) (*models.AIEvaluationRun, error)
}

type aiEvaluationRepository struct {
	db *gorm.DB
}

func NewAIEvaluationRepository(db *gorm.DB) AIEvaluationRepository {
	return &aiEvaluationRepository{db: db}
}

func (r *aiEvaluationRepository) CreateRun(ctx context.Context, run *models.AIEvaluationRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *aiEvaluationRepository) UpdateRun(ctx context.Context, run *models.AIEvaluationRun) error {
	return r.db.WithContext(ctx).Save(run).Error
}

func (r *aiEvaluationRepository) SaveDetail(ctx context.Context, detail *models.AIEvaluationResultDetail) error {
	return r.db.WithContext(ctx).Create(detail).Error
}

func (r *aiEvaluationRepository) GetRunByID(ctx context.Context, id int64) (*models.AIEvaluationRun, error) {
	var run models.AIEvaluationRun
	if err := r.db.WithContext(ctx).First(&run, id).Error; err != nil {
		return nil, err
	}
	return &run, nil
}
