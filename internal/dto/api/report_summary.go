package dto_api

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
