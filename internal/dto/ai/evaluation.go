package dto_ai

type EvaluationCase struct {
	CaseID                string                 `json:"case_id"`
	SyntheticInput        ExceptionPromptContext `json:"synthetic_input"`
	ExpectedSeverity      string                 `json:"expected_severity"`
	ExpectedExceptionType string                 `json:"expected_exception_type"`
}

type EvaluationDataset struct {
	Cases []EvaluationCase `json:"cases"`
}

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

type EvaluationDetailItem struct {
	ID             int64   `json:"id"`
	Status         string  `json:"status"`
	LatencyMs      int     `json:"latency_ms"`
	ExpectedOutput string  `json:"expected_output"`
	ActualOutput   string  `json:"actual_output"`
	ErrorMessage   *string `json:"error_message,omitempty"`
}

type GetEvaluationDetailsResponse struct {
	RunID   int64                  `json:"run_id"`
	Status  string                 `json:"status"`
	Details []EvaluationDetailItem `json:"details"`
}
