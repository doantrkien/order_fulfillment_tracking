package models

import (
	"time"

	"gorm.io/datatypes"
)
type AICustomerUpdateDraft struct {
	ID                  int64  `gorm:"primaryKey;column:id" json:"id"`
	OrderID             int64  `gorm:"column:order_id;not null;index" json:"order_id"`
	AIExceptionResultID *int64 `gorm:"column:ai_exception_result_id" json:"ai_exception_result_id,omitempty"`

	// Draft Content
	DraftMessage string `gorm:"column:draft_message;type:text;not null" json:"draft_message"`
	Tone         string `gorm:"column:tone;type:varchar(20);not null;default:'neutral'" json:"tone"`
	Channel      string `gorm:"column:channel;type:varchar(20);not null;default:'email'" json:"channel"`


	// Human Review State
	ReviewStatus string     `gorm:"column:review_status;type:varchar(20);not null;default:'PENDING'" json:"review_status"`
	ReviewedBy   *string    `gorm:"column:reviewed_by;type:varchar(100)" json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time `gorm:"column:reviewed_at" json:"reviewed_at,omitempty"`
	SentAt       *time.Time `gorm:"column:sent_at" json:"sent_at,omitempty"`

	// AI Metadata
	ConfidenceScore       *float64 `gorm:"column:confidence_score;type:decimal(4,3)" json:"confidence_score,omitempty"`
	FallbackUsed          bool     `gorm:"column:fallback_used;not null;default:false" json:"fallback_used"`
	FallbackReason        *string  `gorm:"column:fallback_reason;type:text" json:"fallback_reason,omitempty"`
	PromptTemplateVersion string   `gorm:"column:prompt_template_version;type:varchar(20);not null;default:'v1'" json:"prompt_template_version"`
	InputSizeBytes        int      `gorm:"column:input_size_bytes;not null;default:0" json:"input_size_bytes"`
	DurationMs            *int           `gorm:"column:duration_ms" json:"duration_ms,omitempty"`
	RequestID             *string        `gorm:"column:request_id;type:varchar(100)" json:"request_id,omitempty"`
	RawResponse           datatypes.JSON `gorm:"column:raw_response" json:"raw_response,omitempty"`

	// Timestamps
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (AICustomerUpdateDraft) TableName() string {
	return "ai_customer_update_drafts"
}

// ReviewStatusPending, ReviewStatusApproved, etc. để tránh magic string rải rác trong code.
const (
	DraftReviewStatusPending  = "PENDING"
	DraftReviewStatusApproved = "APPROVED"
	DraftReviewStatusRejected = "REJECTED"
	DraftReviewStatusSent     = "SENT"

	DraftToneNeutral     = "neutral"
	DraftToneApologetic  = "apologetic"
	DraftToneInformative = "informative"
	DraftToneProactive   = "proactive"
)
