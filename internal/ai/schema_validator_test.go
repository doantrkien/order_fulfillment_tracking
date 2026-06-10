package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAndValidateAIOutput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		wantOutput  *AIExceptionRawOutput
	}{
		{
			name: "valid_happy_path",
			input: `{
				"exception_type": "STUCK_ORDER",
				"severity": "HIGH",
				"likely_reason": "Order stuck at packed status for 72 hours",
				"internal_next_action": "Escalate to warehouse team for shipment check",
				"confidence_score": 0.85
			}`,
			wantErr: false,
			wantOutput: &AIExceptionRawOutput{
				ExceptionType:      "STUCK_ORDER",
				Severity:           "HIGH",
				LikelyReason:       "Order stuck at packed status for 72 hours",
				InternalNextAction: "Escalate to warehouse team for shipment check",
				ConfidenceScore:    0.85,
			},
		},
		{
			name: "valid_all_exception_types",
			input: `{
				"exception_type": "INVALID_TRANSITION",
				"severity": "CRITICAL",
				"likely_reason": "Attempted created to delivered skip",
				"internal_next_action": "Review event source",
				"confidence_score": 0.92
			}`,
			wantErr: false,
		},
		{
			name:    "valid_with_markdown_fences",
			input:   "```json\n{\"exception_type\":\"OTHER\",\"severity\":\"LOW\",\"likely_reason\":\"Minor delay\",\"internal_next_action\":\"Monitor\",\"confidence_score\":0.7}\n```",
			wantErr: false,
			wantOutput: &AIExceptionRawOutput{
				ExceptionType:      "OTHER",
				Severity:           "LOW",
				LikelyReason:       "Minor delay",
				InternalNextAction: "Monitor",
				ConfidenceScore:    0.7,
			},
		},
		{
			name: "valid_boundary_confidence_zero",
			input: `{
				"exception_type": "OTHER",
				"severity": "LOW",
				"likely_reason": "Cannot determine",
				"internal_next_action": "Manual review needed",
				"confidence_score": 0.0
			}`,
			wantErr: false,
		},
		{
			name: "valid_boundary_confidence_one",
			input: `{
				"exception_type": "DELIVERY_FAILURE",
				"severity": "CRITICAL",
				"likely_reason": "Driver reported package lost",
				"internal_next_action": "File insurance claim",
				"confidence_score": 1.0
			}`,
			wantErr: false,
		},
		// ── Invalid cases ──────────────────────────────────────────────
		{
			name:        "invalid_not_json",
			input:       "This is not JSON at all",
			wantErr:     true,
			errContains: "not valid JSON",
		},
		{
			name:        "invalid_empty_string",
			input:       "",
			wantErr:     true,
			errContains: "not valid JSON",
		},
		{
			name: "invalid_exception_type",
			input: `{
				"exception_type": "UNKNOWN_TYPE",
				"severity": "HIGH",
				"likely_reason": "Some reason",
				"internal_next_action": "Some action",
				"confidence_score": 0.8
			}`,
			wantErr:     true,
			errContains: "invalid exception_type",
		},
		{
			name: "invalid_severity",
			input: `{
				"exception_type": "STUCK_ORDER",
				"severity": "EXTREME",
				"likely_reason": "Some reason",
				"internal_next_action": "Some action",
				"confidence_score": 0.8
			}`,
			wantErr:     true,
			errContains: "invalid severity",
		},
		{
			name: "invalid_empty_likely_reason",
			input: `{
				"exception_type": "STUCK_ORDER",
				"severity": "HIGH",
				"likely_reason": "",
				"internal_next_action": "Some action",
				"confidence_score": 0.8
			}`,
			wantErr:     true,
			errContains: "likely_reason is required",
		},
		{
			name: "invalid_empty_next_action",
			input: `{
				"exception_type": "STUCK_ORDER",
				"severity": "HIGH",
				"likely_reason": "Some reason",
				"internal_next_action": "",
				"confidence_score": 0.8
			}`,
			wantErr:     true,
			errContains: "internal_next_action is required",
		},
		{
			name: "invalid_likely_reason_too_long",
			input: `{
				"exception_type": "STUCK_ORDER",
				"severity": "HIGH",
				"likely_reason": "` + strings.Repeat("x", 201) + `",
				"internal_next_action": "Some action",
				"confidence_score": 0.8
			}`,
			wantErr:     true,
			errContains: "likely_reason exceeds 200 characters",
		},
		{
			name: "invalid_next_action_too_long",
			input: `{
				"exception_type": "STUCK_ORDER",
				"severity": "HIGH",
				"likely_reason": "Some reason",
				"internal_next_action": "` + strings.Repeat("a", 201) + `",
				"confidence_score": 0.8
			}`,
			wantErr:     true,
			errContains: "internal_next_action exceeds 200 characters",
		},
		{
			name: "invalid_confidence_negative",
			input: `{
				"exception_type": "STUCK_ORDER",
				"severity": "HIGH",
				"likely_reason": "Some reason",
				"internal_next_action": "Some action",
				"confidence_score": -0.1
			}`,
			wantErr:     true,
			errContains: "confidence_score must be between",
		},
		{
			name: "invalid_confidence_above_one",
			input: `{
				"exception_type": "STUCK_ORDER",
				"severity": "HIGH",
				"likely_reason": "Some reason",
				"internal_next_action": "Some action",
				"confidence_score": 1.5
			}`,
			wantErr:     true,
			errContains: "confidence_score must be between",
		},
		{
			name: "invalid_multiple_violations",
			input: `{
				"exception_type": "BAD_TYPE",
				"severity": "BAD_SEV",
				"likely_reason": "",
				"internal_next_action": "",
				"confidence_score": 2.0
			}`,
			wantErr:     true,
			errContains: "invalid exception_type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ParseAndValidateAIOutput(tc.input)

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Nil(t, output)
			} else {
				require.NoError(t, err)
				require.NotNil(t, output)

				if tc.wantOutput != nil {
					assert.Equal(t, tc.wantOutput.ExceptionType, output.ExceptionType)
					assert.Equal(t, tc.wantOutput.Severity, output.Severity)
					assert.Equal(t, tc.wantOutput.LikelyReason, output.LikelyReason)
					assert.Equal(t, tc.wantOutput.InternalNextAction, output.InternalNextAction)
					assert.InDelta(t, tc.wantOutput.ConfidenceScore, output.ConfidenceScore, 0.001)
				}
			}
		})
	}
}

func TestParseAndValidateAIOutput_MultipleViolationsCollected(t *testing.T) {
	input := `{
		"exception_type": "INVALID",
		"severity": "WRONG",
		"likely_reason": "",
		"internal_next_action": "",
		"confidence_score": -1.0
	}`

	_, err := ParseAndValidateAIOutput(input)
	require.Error(t, err)

	valErr, ok := err.(*ValidationError)
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
			got := stripMarkdownFences(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}
