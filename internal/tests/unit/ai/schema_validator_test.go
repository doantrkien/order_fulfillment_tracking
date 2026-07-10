package ai

import (
	"main/internal/ai"
	dto_ai "main/internal/dto/ai"
	"main/utils/helpers"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAndValidateAIOutput(t *testing.T) {
	tests := []struct {
		name        string
		input       dto_ai.AIAnalysisResult
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_happy_path",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "STUCK_ORDER",
				Severity:           "HIGH",
				LikelyReason:       "Order stuck at packed status for 72 hours",
				InternalNextAction: "Escalate to warehouse team for shipment check",
				ConfidenceScore:    0.85,
			},
			wantErr: false,
		},
		{
			name: "valid_all_exception_types",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "INVALID_TRANSITION",
				Severity:           "CRITICAL",
				LikelyReason:       "Attempted created to delivered skip",
				InternalNextAction: "Review event source",
				ConfidenceScore:    0.92,
			},
			wantErr: false,
		},
		{
			name: "valid_boundary_confidence_zero",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "OTHER",
				Severity:           "LOW",
				LikelyReason:       "Cannot determine",
				InternalNextAction: "Manual review needed",
				ConfidenceScore:    0.0,
			},
			wantErr: false,
		},
		{
			name: "valid_boundary_confidence_one",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "DELIVERY_FAILURE",
				Severity:           "CRITICAL",
				LikelyReason:       "Driver reported package lost",
				InternalNextAction: "File insurance claim",
				ConfidenceScore:    1.0,
			},
			wantErr: false,
		},
		{
			name: "valid_alternative_delivery",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "ALTERNATIVE_DELIVERY",
				Severity:           "LOW",
				LikelyReason:       "Package left at reception desk per customer arrangement",
				InternalNextAction: "Record alternative handoff location. Flag for confirmation if no pickup within 24h.",
				ConfidenceScore:    0.95,
			},
			wantErr: false,
		},
		// ── Invalid cases ──────────────────────────────────────────────
		{
			name: "invalid_exception_type",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "UNKNOWN_TYPE",
				Severity:           "HIGH",
				LikelyReason:       "Some reason",
				InternalNextAction: "Some action",
				ConfidenceScore:    0.8,
			},
			wantErr:     true,
			errContains: "invalid exception_type",
		},
		{
			name: "invalid_severity",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "STUCK_ORDER",
				Severity:           "EXTREME",
				LikelyReason:       "Some reason",
				InternalNextAction: "Some action",
				ConfidenceScore:    0.8,
			},
			wantErr:     true,
			errContains: "invalid severity",
		},
		{
			name: "invalid_empty_likely_reason",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "STUCK_ORDER",
				Severity:           "HIGH",
				LikelyReason:       "",
				InternalNextAction: "Some action",
				ConfidenceScore:    0.8,
			},
			wantErr:     true,
			errContains: "likely_reason is required",
		},
		{
			name: "invalid_empty_next_action",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "STUCK_ORDER",
				Severity:           "HIGH",
				LikelyReason:       "Some reason",
				InternalNextAction: "",
				ConfidenceScore:    0.8,
			},
			wantErr:     true,
			errContains: "internal_next_action is required",
		},
		{
			name: "invalid_likely_reason_too_long",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "STUCK_ORDER",
				Severity:           "HIGH",
				LikelyReason:       strings.Repeat("x", 201),
				InternalNextAction: "Some action",
				ConfidenceScore:    0.8,
			},
			wantErr:     true,
			errContains: "likely_reason exceeds 200 characters",
		},
		{
			name: "invalid_next_action_too_long",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "STUCK_ORDER",
				Severity:           "HIGH",
				LikelyReason:       "Some reason",
				InternalNextAction: strings.Repeat("a", 201),
				ConfidenceScore:    0.8,
			},
			wantErr:     true,
			errContains: "internal_next_action exceeds 200 characters",
		},
		{
			name: "invalid_confidence_negative",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "STUCK_ORDER",
				Severity:           "HIGH",
				LikelyReason:       "Some reason",
				InternalNextAction: "Some action",
				ConfidenceScore:    -0.1,
			},
			wantErr:     true,
			errContains: "confidence_score must be between",
		},
		{
			name: "invalid_confidence_above_one",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "STUCK_ORDER",
				Severity:           "HIGH",
				LikelyReason:       "Some reason",
				InternalNextAction: "Some action",
				ConfidenceScore:    1.5,
			},
			wantErr:     true,
			errContains: "confidence_score must be between",
		},
		{
			name: "invalid_multiple_violations",
			input: dto_ai.AIAnalysisResult{
				ExceptionType:      "BAD_TYPE",
				Severity:           "BAD_SEV",
				LikelyReason:       "",
				InternalNextAction: "",
				ConfidenceScore:    2.0,
			},
			wantErr:     true,
			errContains: "invalid exception_type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ai.ParseAndValidateAIOutput(&tc.input)

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidateAIOutput_MultipleViolationsCollected(t *testing.T) {
	input := dto_ai.AIAnalysisResult{
		ExceptionType:      "INVALID",
		Severity:           "WRONG",
		LikelyReason:       "",
		InternalNextAction: "",
		ConfidenceScore:    -1.0,
	}

	err := ai.ParseAndValidateAIOutput(&input)
	require.Error(t, err)

	valErr, ok := err.(*ai.ValidationError)
	require.True(t, ok, "error should be *ValidationError")
	assert.GreaterOrEqual(t, len(valErr.Fields), 4, "should collect multiple violations")
}

func TestStripMarkdownFences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no_fences", `{"key":"value"}`, `{"key":"value"}`},
		{"json_fences", "```json\n{\"key\":\"value\"}\n```", `{"key":"value"}`},
		{"plain_fences", "```\n{\"key\":\"value\"}\n```", `{"key":"value"}`},
		{"with_whitespace", "  ```json\n{\"key\":\"value\"}\n```  ", `{"key":"value"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := helpers.StripFences(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}
