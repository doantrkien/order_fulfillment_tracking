package ai

import (
	"main/internal/dto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldFallback(t *testing.T) {
	tests := []struct {
		name         string
		confidence   float64
		wantFallback bool
		wantReason   string
	}{
		{
			name:         "below_threshold_0.0",
			confidence:   0.0,
			wantFallback: true,
			wantReason:   "ai_confidence_below_threshold",
		},
		{
			name:         "below_threshold_0.59",
			confidence:   0.59,
			wantFallback: true,
			wantReason:   "ai_confidence_below_threshold",
		},
		{
			name:         "at_threshold_0.6_accept",
			confidence:   0.6,
			wantFallback: false,
			wantReason:   "",
		},
		{
			name:         "above_threshold_0.61",
			confidence:   0.61,
			wantFallback: false,
			wantReason:   "",
		},
		{
			name:         "high_confidence_0.95",
			confidence:   0.95,
			wantFallback: false,
			wantReason:   "",
		},
		{
			name:         "max_confidence_1.0",
			confidence:   1.0,
			wantFallback: false,
			wantReason:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := &dto.ExceptionOutput{
				ExceptionType:      "STUCK_ORDER",
				Severity:           "HIGH",
				LikelyReason:       "test reason",
				InternalNextAction: "test action",
				ConfidenceScore:    tc.confidence,
			}

			gotFallback, gotReason := ShouldFallback(output)
			assert.Equal(t, tc.wantFallback, gotFallback)
			assert.Equal(t, tc.wantReason, gotReason)
		})
	}
}
