package validates

import dto_ai "main/internal/dto/ai"

const ConfidenceThreshold = 0.6

func ShouldFallback(output *dto_ai.AIAnalysisResult) (bool, string) {
	if output.ConfidenceScore < ConfidenceThreshold {
		return true, "ai_confidence_below_threshold"
	}
	return false, ""
}
