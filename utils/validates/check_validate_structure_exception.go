package validates

func IsStructuralException(exceptionType string) bool {
	switch exceptionType {
	case "INVALID_TRANSITION",
		"SKIPPED_STATUS",
		"DUPLICATE_EVENT",
		"CANCELLATION_ANOMALY",
		"REFUND_ANOMALY",
		"NONE":
		return true
	}
	return false
}
