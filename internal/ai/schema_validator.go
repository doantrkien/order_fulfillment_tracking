package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

type AIExceptionRawOutput struct {
	// ExceptionType      string  `json:"exception_type"`
	// Severity           string  `json:"severity"`
	// LikelyReason       string  `json:"likely_reason"`
	// InternalNextAction string  `json:"internal_next_action"`
	// ShouldAlert        bool    `json:"should_alert"`
	// ConfidenceScore    float64 `json:"confidence_score"`
	ExceptionType      string  `json:"exception_type"`
	Severity           string  `json:"severity"`
	LikelyReason       string  `json:"likely_reason"`
	InternalNextAction string  `json:"internal_next_action"`
	Suggestion         string  `json:"suggestion"`
	ShouldAlert        bool    `json:"should_alert"`
	ConfidenceScore    float64 `json:"confidence_score"`
}

var validExceptionTypes = map[string]bool{
	"INVALID_TRANSITION":   true,
	"STUCK_ORDER":          true,
	"SKIPPED_STATUS":       true,
	"DUPLICATE_EVENT":      true,
	"DELIVERY_FAILURE":     true,
	"CANCELLATION_ANOMALY": true,
	"REFUND_ANOMALY":       true,
	"OTHER":                true,
}

var validSeverities = map[string]bool{
	"LOW":      true,
	"MEDIUM":   true,
	"HIGH":     true,
	"CRITICAL": true,
}

type ValidationError struct {
	Fields []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("schema validation failed: %s", strings.Join(e.Fields, "; "))
}

func ParseAndValidateAIOutput(rawResponse string) (*AIExceptionRawOutput, error) {
	cleaned := stripMarkdownFences(rawResponse)

	var output AIExceptionRawOutput
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return nil, fmt.Errorf("AI response is not valid JSON: %w", err)
	}

	var violations []string

	if !validExceptionTypes[output.ExceptionType] {
		violations = append(violations,
			fmt.Sprintf("invalid exception_type '%s' (expected one of: %s)",
				output.ExceptionType, joinMapKeys(validExceptionTypes)))
	}

	if !validSeverities[output.Severity] {
		violations = append(violations,
			fmt.Sprintf("invalid severity '%s' (expected one of: LOW, MEDIUM, HIGH, CRITICAL)",
				output.Severity))
	}

	if output.LikelyReason == "" {
		violations = append(violations, "likely_reason is required and cannot be empty")
	} else if len(output.LikelyReason) > 200 {
		violations = append(violations,
			fmt.Sprintf("likely_reason exceeds 200 characters (got %d)", len(output.LikelyReason)))
	}

	if output.InternalNextAction == "" {
		violations = append(violations, "internal_next_action is required and cannot be empty")
	} else if len(output.InternalNextAction) > 200 {
		violations = append(violations,
			fmt.Sprintf("internal_next_action exceeds 200 characters (got %d)", len(output.InternalNextAction)))
	}

	if output.ConfidenceScore < 0.0 || output.ConfidenceScore > 1.0 {
		violations = append(violations,
			fmt.Sprintf("confidence_score must be between 0.0 and 1.0 (got %.4f)", output.ConfidenceScore))
	}

	if len(violations) > 0 {
		return nil, &ValidationError{Fields: violations}
	}

	return &output, nil
}

func stripMarkdownFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func joinMapKeys(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}
