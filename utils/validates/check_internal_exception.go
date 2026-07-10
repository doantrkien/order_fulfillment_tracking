package validates

func IsInternalSystemError(exceptionType string) bool {
	switch exceptionType {
	case "INVALID_TRANSITION", "SKIPPED_STATUS", "DUPLICATE_EVENT":
		return true
	}
	return false
}
