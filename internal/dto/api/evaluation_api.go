package dto_api

type TriggerEvaluationRequest struct {
	DatasetName string `json:"dataset_name" validate:"required" example:"evaluation_cases.json"`
}

type TriggerEvaluationResponse struct {
	RunID int64 `json:"run_id"`
	// Status  string `json:"status"`
	Message string `json:"message"`
}
