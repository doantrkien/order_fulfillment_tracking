package repositories

import (
	"context"
	"main/internal/models"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// ChunkWithEntry holds a chunk together with its parent entry metadata.
// It is returned by FindSimilarChunks after joining knowledge_chunks with knowledge_entries.
type ChunkWithEntry struct {
	ChunkID    int64   `gorm:"column:chunk_id"`
	EntrySlug  string  `gorm:"column:entry_slug"`
	EntryTitle string  `gorm:"column:entry_title"`
	Heading    string  `gorm:"column:heading"`
	Content    string  `gorm:"column:content"`
	Similarity float64 `gorm:"column:similarity"`
}

type KnowledgeRepository interface {
	GetAllActive(ctx context.Context) ([]models.KnowledgeEntry, error)
	GetBySlug(ctx context.Context, slug string) (*models.KnowledgeEntry, error)
	Save(ctx context.Context, entry *models.KnowledgeEntry) error
	FindSimilar(ctx context.Context, vec []float32, topK int, threshold float64) ([]models.KnowledgeEntry, error)
	UpdateEmbedding(ctx context.Context, id int64, vec []float32) error
	SaveChunks(ctx context.Context, chunks []models.KnowledgeChunk) error
	DeleteChunksByEntryID(ctx context.Context, entryID int64) error
	FindSimilarChunks(ctx context.Context, vec []float32, topK int, threshold float64) ([]ChunkWithEntry, error)
	UpdateChunkEmbedding(ctx context.Context, chunkID int64, vec []float32) error
	GetChunksNeedingReembed(ctx context.Context) ([]models.KnowledgeChunk, error)
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

func (r *knowledgeRepository) SaveChunks(ctx context.Context, chunks []models.KnowledgeChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Omit("Embedding").
		Create(&chunks).Error
}

func (r *knowledgeRepository) DeleteChunksByEntryID(ctx context.Context, entryID int64) error {
	return r.db.WithContext(ctx).
		Where("entry_id = ?", entryID).
		Delete(&models.KnowledgeChunk{}).Error
}

func (r *knowledgeRepository) FindSimilarChunks(ctx context.Context, vec []float32, topK int, threshold float64) ([]ChunkWithEntry, error) {
	pgVec := pgvector.NewVector(vec)
	var results []ChunkWithEntry
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT
				kc.id AS chunk_id,
				ke.slug AS entry_slug,
				ke.title AS entry_title,
				kc.heading,
				kc.content,
				1 - (kc.embedding <=> ?) AS similarity
			FROM knowledge_chunks kc
			JOIN knowledge_entries ke ON kc.entry_id = ke.id
			WHERE ke.is_active = true
			  AND kc.embedding IS NOT NULL
			  AND 1 - (kc.embedding <=> ?) > ?
			ORDER BY similarity DESC
			LIMIT ?
		`, pgVec, pgVec, threshold, topK).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *knowledgeRepository) UpdateChunkEmbedding(ctx context.Context, chunkID int64, vec []float32) error {
	pgVec := pgvector.NewVector(vec)
	return r.db.WithContext(ctx).
		Model(&models.KnowledgeChunk{}).
		Where("id = ?", chunkID).
		Updates(map[string]any{
			"embedding":     pgVec,
			"needs_reembed": false,
		}).Error
}

func (r *knowledgeRepository) GetChunksNeedingReembed(ctx context.Context) ([]models.KnowledgeChunk, error) {
	var chunks []models.KnowledgeChunk
	err := r.db.WithContext(ctx).
		Where("needs_reembed = ?", true).
		Find(&chunks).Error
	if err != nil {
		return nil, err
	}
	return chunks, nil
}
