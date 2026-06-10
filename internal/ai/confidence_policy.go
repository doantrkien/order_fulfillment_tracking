package ai

const ConfidenceThreshold = 0.6

func ShouldFallback(output *AIExceptionRawOutput) (bool, string) {
	if output.ConfidenceScore < ConfidenceThreshold {
		return true, "ai_confidence_below_threshold"
	}
	return false, ""
}
