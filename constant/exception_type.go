package constant

var ValidExceptionTypes = map[string]bool{
	"INVALID_TRANSITION":   true,
	"CANCELLATION_ANOMALY": true,
	"REFUND_ANOMALY":       true,
	"STUCK_ORDER":          true,
	"SKIPPED_STATUS":       true,
	"DUPLICATE_EVENT":      true,
	"DELIVERY_FAILURE":     true,
	"ALTERNATIVE_DELIVERY": true,
	"OTHER":                true,
}
