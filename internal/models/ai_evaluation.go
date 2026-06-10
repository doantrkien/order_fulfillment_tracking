package models

import (
	"time"

	"gorm.io/datatypes"
)

type AIEvaluationRun struct {
	ID int64 `gorm:"primaryKey;column:id" json:"id"`

	// Run Metadata
	WorkflowName          string  `gorm:"column:workflow_name;type:varchar(100);not null" json:"workflow_name"`
	PromptTemplateVersion string  `gorm:"column:prompt_template_version;type:varchar(20);not null;default:'v1'" json:"prompt_template_version"`
	RunBy                 *string `gorm:"column:run_by;type:varchar(100)" json:"run_by,omitempty"`
	Environment           string  `gorm:"column:environment;type:varchar(20);not null;default:'dev'" json:"environment"`
	TriggeredBy           string  `gorm:"column:triggered_by;type:varchar(50);not null;default:'manual'" json:"triggered_by"`

	// Aggregate Results
	TotalCases    int      `gorm:"column:total_cases;not null;default:0" json:"total_cases"`
	PassedCases   int      `gorm:"column:passed_cases;not null;default:0" json:"passed_cases"`
	FailedCases   int      `gorm:"column:failed_cases;not null;default:0" json:"failed_cases"`
	FallbackCases int      `gorm:"column:fallback_cases;not null;default:0" json:"fallback_cases"`
	AvgConfidence *float64 `gorm:"column:avg_confidence;type:decimal(4,3)" json:"avg_confidence,omitempty"`
	PassRate      *float64 `gorm:"column:pass_rate;type:decimal(5,4)" json:"pass_rate,omitempty"`

	// Detail & Debug
	Notes      *string        `gorm:"column:notes;type:text" json:"notes,omitempty"`
	RawResults datatypes.JSON `gorm:"column:raw_results;type:jsonb" json:"raw_results,omitempty"`

	// Timestamps
	RunAt time.Time `gorm:"column:run_at;not null;default:CURRENT_TIMESTAMP" json:"run_at"`
}

func (AIEvaluationRun) TableName() string {
	return "ai_evaluation_runs"
}

const (
	EvalEnvironmentDev     = "dev"
	EvalEnvironmentStaging = "staging"
	EvalEnvironmentProd    = "prod"

	EvalTriggeredByManual    = "manual"
	EvalTriggeredByCI        = "ci"
	EvalTriggeredByScheduled = "scheduled"

	// WorkflowNameOrderException là tên chuẩn dùng cho Team 2.
	WorkflowNameOrderException = "order-exception-analysis"
)
