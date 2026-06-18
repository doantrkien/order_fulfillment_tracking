package dto

import "time"

type AnalyzeExceptionRequest struct {
	Note string `json:"note" example:"Phân tích đơn hàng này hộ tôi"`
}

type AnalyzeExceptionResponse struct {
	ResultID              string    `json:"result_id"`
	OrderID               string    `json:"order_id"`
	ExceptionType         string    `json:"exception_type"`
	Severity              string    `json:"severity"`
	LikelyReason          string    `json:"likely_reason"`
	InternalNextAction    string    `json:"internal_next_action"`
	CustomerUpdateDraft   string    `json:"customer_update_draft"`
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

type UpdateDraftAPIRequest struct {
	OrderID int64  `json:"order_id"`
	Tone    string `json:"tone"`
	Channel string `json:"channel"`
}

type UpdateDraftAPIResponse struct {
	OrderID               int64     `json:"order_id"`
	DraftMessage          string    `json:"draft_message"`
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

type CustomerUpdateDraftInput struct {
	OrderID         int64  `json:"order_id"`
	CustomerName    string `json:"customer_name"`
	ShippingAddress string `json:"shipping_address"`
	CurrentStatus   string `json:"current_status"`
	ExceptionType   string `json:"exception_type"`
	LikelyReason    string `json:"likely_reason"`
	Tone            string `json:"tone"`
	Channel         string `json:"channel"`
}

type CustomerUpdateDraftOutput struct {
	CustomerUpdateDraft string  `json:"customer_update_draft"`
	ConfidenceScore     float64 `json:"confidence_score"`
}

type EvaluationCase struct {
	CaseID                string          `json:"case_id"`
	OrderID               int64           `json:"order_id,omitempty"`
	SyntheticInput        *ExceptionInput `json:"synthetic_input,omitempty"`
	ExpectedSeverity      string          `json:"expected_severity"`
	ExpectedExceptionType string          `json:"expected_exception_type"`
}

type EvaluationRequest struct {
	RunBy       string           `json:"run_by"`
	Environment string           `json:"environment,omitempty"`
	Cases       []EvaluationCase `json:"cases"`
}

type EvaluationCaseResult struct {
	CaseID              string  `json:"case_id"`
	Passed              bool    `json:"passed"`
	FallbackUsed        bool    `json:"fallback_used"`
	ActualSeverity      string  `json:"actual_severity"`
	ActualExceptionType string  `json:"actual_exception_type"`
	ConfidenceScore     float64 `json:"confidence_score"`
	FailReason          string  `json:"fail_reason,omitempty"`
}

type EvaluationResponse struct {
	WorkflowName  string                 `json:"workflow_name"`
	TotalCases    int                    `json:"total_cases"`
	PassedCases   int                    `json:"passed_cases"`
	FailedCases   int                    `json:"failed_cases"`
	FallbackCases int                    `json:"fallback_cases"`
	PassRate      float64                `json:"pass_rate"`
	AvgConfidence float64                `json:"avg_confidence"`
	Results       []EvaluationCaseResult `json:"results"`
	RunAt         time.Time              `json:"run_at"`
}
