package repositories

import (
	"context"
	"main/internal/models"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type KnowledgeRepository interface {
	GetAllActive(ctx context.Context) ([]models.KnowledgeEntry, error)
	GetBySlug(ctx context.Context, slug string) (*models.KnowledgeEntry, error)
	Save(ctx context.Context, entry *models.KnowledgeEntry) error
	// FindSimilar returns up to topK entries whose embedding cosine-similarity with
	// vec is >= threshold. Results are ordered by similarity descending.
	FindSimilar(ctx context.Context, vec []float32, topK int, threshold float64) ([]models.KnowledgeEntry, error)
	// UpdateEmbedding persists the embedding vector for a single entry and clears
	// the needs_reembed flag.
	UpdateEmbedding(ctx context.Context, id int64, vec []float32) error
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
		Omit("Embedding").
		Save(entry).Error
}

// FindSimilar queries knowledge_entries using pgvector cosine distance operator (<=>).
// Only entries with a non-NULL embedding and is_active=true are searched.
// Similarity is defined as 1 - cosine_distance, so threshold=0.75 means cosine distance <= 0.25.
func (r *knowledgeRepository) FindSimilar(ctx context.Context, vec []float32, topK int, threshold float64) ([]models.KnowledgeEntry, error) {
	pgVec := pgvector.NewVector(vec)
	var entries []models.KnowledgeEntry
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT id, slug, title, body, is_active, needs_reembed, updated_at
			FROM knowledge_entries
			WHERE is_active = true
			  AND embedding IS NOT NULL
			  AND 1 - (embedding <=> ?) > ?
			ORDER BY embedding <=> ?
			LIMIT ?
		`, pgVec, threshold, pgVec, topK).
		Scan(&entries).Error
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// UpdateEmbedding saves the embedding vector for the given entry ID and resets needs_reembed.
func (r *knowledgeRepository) UpdateEmbedding(ctx context.Context, id int64, vec []float32) error {
	pgVec := pgvector.NewVector(vec)
	return r.db.WithContext(ctx).
		Model(&models.KnowledgeEntry{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"embedding":     pgVec,
			"needs_reembed": false,
		}).Error
}
