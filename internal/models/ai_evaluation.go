package models

import (
	"time"

	"gorm.io/datatypes"
)

// Evaluation run status constants.
const (
	EVAL_STATUS_PENDING     = "PENDING"
	EVAL_STATUS_IN_PROGRESS = "IN_PROGRESS"
	EVAL_STATUS_COMPLETED   = "COMPLETED"
	EVAL_STATUS_FAILED      = "FAILED"
)

// Evaluation detail status constants.
const (
	EVAL_DETAIL_PENDING = "PENDING"
	EVAL_DETAIL_PASSED  = "PASSED"
	EVAL_DETAIL_FAILED  = "FAILED"
)

// AIEvaluationRun represents a batch evaluation run record.
type AIEvaluationRun struct {
	ID            int64     `gorm:"primaryKey;column:id" json:"id"`
	DatasetName   string    `gorm:"column:dataset_name;type:varchar(255);not null" json:"dataset_name"`
	Status        string    `gorm:"column:status;type:varchar(20);not null;default:'PENDING'" json:"status"`
	TotalCases    int       `gorm:"column:total_cases;not null;default:0" json:"total_cases"`
	PassedCases   int       `gorm:"column:passed_cases;not null;default:0" json:"passed_cases"`
	FailedCases   int       `gorm:"column:failed_cases;not null;default:0" json:"failed_cases"`
	FallbackCount int       `gorm:"column:fallback_count;not null;default:0" json:"fallback_count"`
	AccuracyRate  float64   `gorm:"column:accuracy_rate;type:decimal(5,2);default:0" json:"accuracy_rate"`
	AvgLatencyMs  int       `gorm:"column:avg_latency_ms;default:0" json:"avg_latency_ms"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (AIEvaluationRun) TableName() string {
	return "ai_evaluation_runs"
}

// AIEvaluationResultDetail represents a single test case result.
type AIEvaluationResultDetail struct {
	ID             int64          `gorm:"primaryKey;column:id" json:"id"`
	RunID          int64          `gorm:"column:run_id;not null;index" json:"run_id"`
	Input          datatypes.JSON `gorm:"column:input;type:jsonb;not null" json:"input"`
	ExpectedOutput datatypes.JSON `gorm:"column:expected_output;type:jsonb;not null" json:"expected_output"`
	ActualOutput   datatypes.JSON `gorm:"column:actual_output;type:jsonb" json:"actual_output"`
	Status         string         `gorm:"column:status;type:varchar(20);not null;default:'PENDING'" json:"status"`
	LatencyMs      int            `gorm:"column:latency_ms;default:0" json:"latency_ms"`
	ErrorMessage   *string        `gorm:"column:error_message;type:text" json:"error_message,omitempty"`
	CreatedAt      time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (AIEvaluationResultDetail) TableName() string {
	return "ai_evaluation_results_detail"
}
