package ai

import "main/internal/dto"

const ConfidenceThreshold = 0.6

func ShouldFallback(output *dto.ExceptionOutput) (bool, string) {
	if output.ConfidenceScore < ConfidenceThreshold {
		return true, "ai_confidence_below_threshold"
	}
	return false, ""
}
