package constant

import "main/internal/models"

var ValidSeverities = map[string]bool{
	"LOW":      true,
	"MEDIUM":   true,
	"HIGH":     true,
	"CRITICAL": true,
}

var CancellationSeverity = map[models.OrderStatus]string{
	models.ORDER_STATUS_CREATED: "LOW",
}

// refundSeverity maps the last meaningful status before refund to a severity level.
var RefundSeverity = map[models.OrderStatus]string{
	models.ORDER_STATUS_PAID:      "LOW",
	models.ORDER_STATUS_PACKED:    "MEDIUM",
	models.ORDER_STATUS_SHIPPED:   "HIGH",
	models.ORDER_STATUS_DELIVERED: "CRITICAL",
}
