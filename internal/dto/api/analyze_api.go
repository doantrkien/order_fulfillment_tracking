package dto_api

import "time"

type AnalyzeExceptionRequest struct {
	Note string `json:"note" validate:"max=500" example:"Phân tích đơn hàng này hộ tôi"`
}

type AnalyzeExceptionResponse struct {
	ResultID              string    `json:"result_id"`
	OrderID               string    `json:"order_id"`
	ExceptionType         string    `json:"exception_type"`
	Severity              string    `json:"severity"`
	LikelyReason          string    `json:"likely_reason"`
	InternalNextAction    string    `json:"internal_next_action"`
	ConfidenceScore       float64   `json:"confidence_score"`
	FallbackUsed          bool      `json:"fallback_used"`
	PromptTemplateVersion string    `json:"prompt_template_version"`
	EvaluatedAt           time.Time `json:"evaluated_at"`
}
