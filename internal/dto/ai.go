package dto

import "time"

type AnalyzeExceptionRequest struct {
	Note string `json:"note"`
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
	FallbackReason        string    `json:"fallback_reason,omitempty"`
	PromptTemplateVersion string    `json:"prompt_template_version"`
	EvaluatedAt           time.Time `json:"evaluated_at"`
}

type EventRecord struct {
	FromStatus string `json:"from_status"`
	ToStatus   string `json:"to_status"`
	EventAt    string `json:"event_at"`
	UpdatedBy  string `json:"updated_by"`
}

type ExceptionInput struct {
	OrderID         int64         `json:"order_id"`
	CurrentStatus   string        `json:"current_status"`
	TotalAmount     int64         `json:"total_amount"`
	CustomerName    string        `json:"customer_name"`
	ShippingAddress string        `json:"shipping_address"`
	CreatedAt       string        `json:"created_at"`
	ErrorMessage    string        `json:"error_message"`
	EventHistory    []EventRecord `json:"event_history"`
}

type ExceptionOutput struct {
	ExceptionType      string  `json:"exception_type"`
	Severity           string  `json:"severity"`
	LikelyReason       string  `json:"likely_reason"`
	InternalNextAction string  `json:"internal_next_action"`
	Suggestion         string  `json:"suggestion"`
	ShouldAlert        bool    `json:"should_alert"`
	Confidence         float64 `json:"confidence"`
}

type UpdateDraftAIInput struct {
	OrderID         int64  `json:"order_id"`
	Tone            string `json:"tone"`
	Channel         string `json:"channel"`
	CustomerName    string `json:"customer_name"`
	ShippingAddress string `json:"shipping_address"`
}

type UpdateDraftAPIRequest struct {
	OrderID int64  `json:"order_id"`
	Tone    string `json:"tone"`
	Channel string `json:"channel"`
}

type UpdateDraftAPIResponse struct {
	OrderID               int64     `json:"order_id"`
	DrafMessage           string    `json:"draft_message"`
	Tone                  string    `json:"tone"`
	ConfidenceScore       float64   `json:"confidence_score"`
	FallbackUsed          bool      `json:"fallback_used"`
	FallbackReason        string    `json:"fallback_reason,omitempty"`
	PromptTemplateVersion string    `json:"prompt_template_version"`
	GeneratedAt           time.Time `json:"generated_at"`
}

type ReportSummaryInput struct {
	Date           string  `json:"date"`
	TotalOrders    int64   `json:"total_orders"`
	TotalNew       int64   `json:"total_new"`
	TotalDelivered int64   `json:"total_delivered"`
	TotalCancelled int64   `json:"total_cancelled"`
	TotalRefunded  int64   `json:"total_refunded"`
	TotalIncome    int64   `json:"total_income"`
	AvgDeliverTime float64 `json:"avg_deliver_time"`
}

type ReportSummaryOutput struct {
	Summary     string   `json:"summary"`
	Highlights  []string `json:"highlights"`
	Suggestions []string `json:"suggestions"`
}
