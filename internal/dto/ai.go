package dto

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
	DriverNotes     string        `json:"driver_notes"`
	EventHistory    []EventRecord `json:"event_history"`
}

type ExceptionOutput struct {
	ExceptionType      string  `json:"exception_type"`
	Severity           string  `json:"severity"`
	LikelyReason       string  `json:"likely_reason"`
	InternalNextAction string  `json:"internal_next_action"`
	Suggestion         string  `json:"suggestion"`
	ShouldAlert        bool    `json:"should_alert"`
	ConfidenceScore    float64 `json:"confidence_score"`
}

type GenerateDraftAPIRequest struct {
	OrderID int64  `json:"order_id" example:"12345"`
	Tone    string `json:"tone" example:"apologetic"`
	Channel string `json:"channel" example:"email"`
}

type GenerateDraftAPIResponse struct {
	OrderID               int64     `json:"order_id"`
	DraftMessage          string    `json:"draft_message"`
	Tone                  string    `json:"tone"`
	ConfidenceScore       float64   `json:"confidence_score"`
	FallbackUsed          bool      `json:"fallback_used"`
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
	BaselineDraft   string `json:"baseline_draft"`
}

type CustomerUpdateDraftOutput struct {
	CustomerUpdateDraft string  `json:"customer_update_draft"`
	ConfidenceScore     float64 `json:"confidence_score"`
}

type TriggerEvaluationRequest struct {
	DatasetName string `json:"dataset_name" validate:"required" example:"evaluation_cases.json"`
}

type TriggerEvaluationResponse struct {
	RunID   int64  `json:"run_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type EvaluationCase struct {
	CaseID                string         `json:"case_id"`
	SyntheticInput        ExceptionInput `json:"synthetic_input"`
	ExpectedSeverity      string         `json:"expected_severity"`
	ExpectedExceptionType string         `json:"expected_exception_type"`
}

type EvaluationDataset struct {
	RunBy       string           `json:"run_by"`
	Environment string           `json:"environment"`
	Cases       []EvaluationCase `json:"cases"`
}

// GetEvaluationRunResponse trả về summary của 1 evaluation run.
type GetEvaluationRunResponse struct {
	RunID         int64   `json:"run_id"`
	DatasetName   string  `json:"dataset_name"`
	Status        string  `json:"status"`
	TotalCases    int     `json:"total_cases"`
	PassedCases   int     `json:"passed_cases"`
	FailedCases   int     `json:"failed_cases"`
	FallbackCount int     `json:"fallback_count"`
	AccuracyRate  float64 `json:"accuracy_rate"`
	AvgLatencyMs  int     `json:"avg_latency_ms"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// EvaluationDetailItem trả về kết quả PASS/FAIL của từng test case.
type EvaluationDetailItem struct {
	ID             int64   `json:"id"`
	Status         string  `json:"status"`
	LatencyMs      int     `json:"latency_ms"`
	ExpectedOutput string  `json:"expected_output"`
	ActualOutput   string  `json:"actual_output"`
	ErrorMessage   *string `json:"error_message,omitempty"`
}

// GetEvaluationDetailsResponse chứa danh sách kết quả chi tiết của 1 run.
type GetEvaluationDetailsResponse struct {
	RunID   int64                  `json:"run_id"`
	Status  string                 `json:"status"`
	Details []EvaluationDetailItem `json:"details"`
}
