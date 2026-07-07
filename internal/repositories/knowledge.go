package repositories

import (
	"context"
	"main/internal/models"

	"gorm.io/gorm"
)

type KnowledgeRepository interface {
	GetAllActive(ctx context.Context) ([]models.KnowledgeEntry, error)
	GetBySlug(ctx context.Context, slug string) (*models.KnowledgeEntry, error)
	Save(ctx context.Context, entry *models.KnowledgeEntry) error
}

type knowledgeRepository struct {
	db *gorm.DB
}

func NewKnowledgeRepository(db *gorm.DB) KnowledgeRepository {
	return &knowledgeRepository{db: db}
}

func (r *knowledgeRepository) GetAllActive(ctx context.Context) ([]models.KnowledgeEntry, error) {
	var entries []models.KnowledgeEntry
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&entries).Error
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *knowledgeRepository) GetBySlug(ctx context.Context, slug string) (*models.KnowledgeEntry, error) {
	var entry models.KnowledgeEntry
	err := r.db.WithContext(ctx).
		Where("slug = ? AND is_active = ?", slug, true).
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *knowledgeRepository) Save(ctx context.Context, entry *models.KnowledgeEntry) error {
	return r.db.WithContext(ctx).
		Save(entry).Error
}
