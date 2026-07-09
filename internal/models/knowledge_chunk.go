package models

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

// KnowledgeChunk represents a section-level fragment of a KnowledgeEntry,
// stored with its own embedding vector for fine-grained semantic retrieval.
type KnowledgeChunk struct {
	ID           int64           `gorm:"primaryKey;column:id" json:"id"`
	EntryID      int64           `gorm:"column:entry_id;not null" json:"entry_id"`
	ChunkIndex   int16           `gorm:"column:chunk_index;not null" json:"chunk_index"`
	Heading      string          `gorm:"column:heading;type:varchar(256);not null" json:"heading"`
	Content      string          `gorm:"column:content;type:text;not null" json:"content"`
	TokenCount   int             `gorm:"column:token_count;not null;default:0" json:"token_count"`
	Embedding    pgvector.Vector `gorm:"column:embedding;type:vector(768)" json:"-"`
	NeedsReembed bool            `gorm:"column:needs_reembed;not null;default:true" json:"-"`
	CreatedAt    time.Time       `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (KnowledgeChunk) TableName() string {
	return "knowledge_chunks"
}
