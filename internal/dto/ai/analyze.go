package dto_ai

type ExceptionPromptContext struct {
	OrderID         int64
	CurrentStatus   string
	TotalAmount     int64
	CustomerName    string
	ShippingAddress string
	CreatedAt       string
	AnalyzedAt      string
	EventTimeline   []EventTimelineEntry
	DriverNotes     string
}

type EventTimelineEntry struct {
	FromStatus string
	ToStatus   string
	UpdatedBy  string
	EventAt    string
}

// type ExceptionInput struct {
// 	OrderID         int64         `json:"order_id"`
// 	CurrentStatus   string        `json:"current_status"`
// 	TotalAmount     int64         `json:"total_amount"`
// 	CustomerName    string        `json:"customer_name"`
// 	ShippingAddress string        `json:"shipping_address"`
// 	CreatedAt       string        `json:"created_at"`
// 	DriverNotes     string        `json:"driver_notes"`
// 	EventHistory    []EventRecord `json:"event_history"`
// }

// type EventRecord struct {
// 	FromStatus string `json:"from_status"`
// 	ToStatus   string `json:"to_status"`
// 	EventAt    string `json:"event_at"`
// 	UpdatedBy  string `json:"updated_by"`
// }

type AIAnalysisResult struct {
	ExceptionType      string  `json:"exception_type"`
	Severity           string  `json:"severity"`
	LikelyReason       string  `json:"likely_reason"`
	InternalNextAction string  `json:"internal_next_action"`
	ConfidenceScore    float64 `json:"confidence_score"`
	FallbackUsed       bool    `json:"fallback_used"`
	FallbackReason     string  `json:"fallback_reason"`
	DurationMs         int     `json:"duration_ms"`
	RawResponse        string  `json:"raw_response"`
}
