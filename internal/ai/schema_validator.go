package ai

import (
	"encoding/json"
	"fmt"
	"main/internal/dto"
	"strings"
)

var validExceptionTypes = map[string]bool{
	"INVALID_TRANSITION": true,
	"STUCK_ORDER":        true,
	"SKIPPED_STATUS":     true,
	"DUPLICATE_EVENT":    true,
	"DELIVERY_FAILURE":   true,
	"OTHER":              true,
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

func ParseAndValidateAIOutput(output *dto.ExceptionOutput) error {
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
		return &ValidationError{Fields: violations}
	}

	return nil
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

type AICustomerUpdateDraftRawOutput struct {
	CustomerUpdateDraft string  `json:"customer_update_draft"`
	ConfidenceScore     float64 `json:"confidence_score"`
}

func ParseAndValidateCustomerUpdateDraft(rawResponse string) (*AICustomerUpdateDraftRawOutput, error) {
	cleaned := stripMarkdownFences(rawResponse)

	var output AICustomerUpdateDraftRawOutput
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return nil, fmt.Errorf("AI response is not valid JSON: %w", err)
	}

	var violations []string

	if output.CustomerUpdateDraft == "" {
		violations = append(violations, "customer_update_draft is required and cannot be empty")
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
