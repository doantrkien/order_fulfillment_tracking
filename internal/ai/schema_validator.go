package ai

import (
	"encoding/json"
	"fmt"
	"main/constant"
	dto_ai "main/internal/dto/ai"
	"main/utils/helpers"
	"strings"
)

type ValidationError struct {
	Fields []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("schema validation failed: %s", strings.Join(e.Fields, "; "))
}

func ParseAndValidateAIOutput(output *dto_ai.AIAnalysisResult) error {
	var violations []string

	if !constant.ValidExceptionTypes[output.ExceptionType] {
		violations = append(violations,
			fmt.Sprintf("invalid exception_type '%s' (expected one of: %s)",
				output.ExceptionType, helpers.JoinMapKeys(constant.ValidExceptionTypes)))
	}

	if !constant.ValidSeverities[output.Severity] {
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

type AICustomerUpdateDraftRawOutput struct {
	CustomerUpdateDraft string  `json:"customer_update_draft"`
	ConfidenceScore     float64 `json:"confidence_score"`
}

func ParseAndValidateCustomerUpdateDraft(rawResponse string) (*AICustomerUpdateDraftRawOutput, error) {
	cleaned := helpers.StripFences(rawResponse)

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
