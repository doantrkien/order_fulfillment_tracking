package dto_api

import "time"

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
