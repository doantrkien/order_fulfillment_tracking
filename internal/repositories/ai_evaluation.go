package repositories

import (
	"context"
	"main/internal/models"

	"gorm.io/gorm"
)

type AIEvaluationRepository interface {
	Save(ctx context.Context, run *models.AIEvaluationRun) error
	GetLatestByWorkflow(ctx context.Context, workflowName string) (*models.AIEvaluationRun, error)
	ListByWorkflow(ctx context.Context, workflowName string, limit int) ([]models.AIEvaluationRun, error)
}

type aiEvaluationRepository struct {
	db *gorm.DB
}

func NewAIEvaluationRepository(db *gorm.DB) AIEvaluationRepository {
	return &aiEvaluationRepository{db: db}
}

func (r *aiEvaluationRepository) Save(ctx context.Context, run *models.AIEvaluationRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *aiEvaluationRepository) GetLatestByWorkflow(ctx context.Context, workflowName string) (*models.AIEvaluationRun, error) {
	var run models.AIEvaluationRun
	err := r.db.WithContext(ctx).
		Where("workflow_name = ?", workflowName).
		Order("run_at DESC").
		First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// ListByWorkflow lấy N lần eval gần nhất của một workflow.
// Dùng cho demo Week 12: so sánh pass rate giữa các prompt version.
func (r *aiEvaluationRepository) ListByWorkflow(ctx context.Context, workflowName string, limit int) ([]models.AIEvaluationRun, error) {
	var runs []models.AIEvaluationRun
	err := r.db.WithContext(ctx).
		Where("workflow_name = ?", workflowName).
		Order("run_at DESC").
		Limit(limit).
		Find(&runs).Error
	if err != nil {
		return nil, err
	}
	return runs, nil
}
