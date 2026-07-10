package dto_ai

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
