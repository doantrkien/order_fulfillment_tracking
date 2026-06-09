package models

import (
	"time"

	"gorm.io/datatypes"
)

type AIException struct {
	ID                    int64          `gorm:"primaryKey;column:id" json:"id"`
	OrderID               int64          `gorm:"column:order_id;not null;index" json:"order_id"`
	ExceptionType         string         `gorm:"column:exception_type;type:varchar(50);not null" json:"exception_type"`
	Severity              string         `gorm:"column:severity;type:varchar(20);not null" json:"severity"`
	LikelyReason          string         `gorm:"column:likely_reason;type:text;not null" json:"likely_reason"`
	InternalNextAction    string         `gorm:"column:internal_next_action;type:text;not null" json:"internal_next_action"`
	ConfidenceScore       float64        `gorm:"column:confidence_score;type:decimal(4,3);not null" json:"confidence_score"`
	FallbackUsed          bool           `gorm:"column:fallback_used;not null;default:false" json:"fallback_used"`
	FallbackReason        *string        `gorm:"column:fallback_reason;type:text" json:"fallback_reason,omitempty"`
	PromptTemplateVersion string         `gorm:"column:prompt_template_version;type:varchar(20);not null;default:'v1'" json:"prompt_template_version"`
	InputSizeBytes        int            `gorm:"column:input_size_bytes;not null;default:0" json:"input_size_bytes"`
	DurationMs            *int           `gorm:"column:duration_ms" json:"duration_ms,omitempty"`
	RequestID             *string        `gorm:"column:request_id;type:varchar(100)" json:"request_id,omitempty"`
	RawResponse           datatypes.JSON `gorm:"column:raw_response;type:jsonb" json:"raw_response,omitempty"`
	EvaluatedAt           time.Time      `gorm:"column:evaluated_at;not null;default:CURRENT_TIMESTAMP" json:"evaluated_at"`
	CreatedAt             time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (AIException) TableName() string {
	return "ai_order_exception_results"
}
