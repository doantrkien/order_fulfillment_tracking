package models

import "time"

type KnowledgeEntry struct {
	ID        int64     `gorm:"primaryKey;column:id" json:"id"`
	Slug      string    `gorm:"column:slug;type:varchar(64);unique;not null" json:"slug"`
	Title     string    `gorm:"column:title;type:varchar(256);not null" json:"title"`
	Body      string    `gorm:"column:body;type:text;not null" json:"body"`
	IsActive  bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (KnowledgeEntry) TableName() string {
	return "knowledge_entries"
}
