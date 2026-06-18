package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"main/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDraftGenerator_AIDisabled(t *testing.T) {
	generator := NewDraftGenerator(nil, DraftGeneratorConfig{
		AIEnabled: false,
		AITimeout: 10 * time.Second,
	})

	input := dto.CustomerUpdateDraftInput{
		OrderID:       1001,
		ExceptionType: "STUCK_ORDER",
		Tone:          "neutral",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, FallbackReasonDisabled, result.FallbackReason)
	assert.Contains(t, result.CustomerUpdateDraft, "slower than expected")
	assert.Equal(t, 1.0, result.ConfidenceScore)
}

func TestDraftGenerator_AIReturnsError(t *testing.T) {
	adapter := &mockAdapter{
		err: errors.New("connection reset by peer"),
	}
	generator := NewDraftGenerator(adapter, DraftGeneratorConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	input := dto.CustomerUpdateDraftInput{
		OrderID:       1002,
		ExceptionType: "DELIVERY_FAILURE",
		Tone:          "apologetic",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err) // DraftGenerator should never return error on AI failure
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, FallbackReasonConnectionError, result.FallbackReason)
	assert.Contains(t, result.CustomerUpdateDraft, "address [REDACTED_SHIPPING_ADDRESS]")
	assert.Equal(t, 1.0, result.ConfidenceScore)
}

func TestDraftGenerator_AIReturnsTimeout(t *testing.T) {
	adapter := &mockAdapter{
		err: context.DeadlineExceeded,
	}
	generator := NewDraftGenerator(adapter, DraftGeneratorConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	input := dto.CustomerUpdateDraftInput{
		OrderID:       1003,
		ExceptionType: "INVALID_TRANSITION",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, FallbackReasonTimeout, result.FallbackReason)
	assert.Contains(t, result.CustomerUpdateDraft, "mismatch in your order status update")
	assert.Equal(t, 1.0, result.ConfidenceScore)
}

func TestDraftGenerator_AIReturnsInvalidResponse(t *testing.T) {
	adapter := &mockAdapter{
		outputStr: "{invalid-json}",
	}
	generator := NewDraftGenerator(adapter, DraftGeneratorConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	input := dto.CustomerUpdateDraftInput{
		OrderID:       1004,
		ExceptionType: "OTHER",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, FallbackReasonInvalidResponse, result.FallbackReason)
	assert.Contains(t, result.CustomerUpdateDraft, "unexpected issue related to your order")
	assert.Equal(t, 1.0, result.ConfidenceScore)
}

func TestDraftGenerator_AIReturnsLowConfidence(t *testing.T) {
	adapter := &mockAdapter{
		outputStr: `{
			"customer_update_draft": "Apology message...",
			"confidence_score": 0.3
		}`,
	}
	generator := NewDraftGenerator(adapter, DraftGeneratorConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	input := dto.CustomerUpdateDraftInput{
		OrderID:       1005,
		ExceptionType: "REFUND_ANOMALY",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, FallbackReasonLowConfidence, result.FallbackReason)
	assert.Contains(t, result.CustomerUpdateDraft, "refund request")
	assert.Equal(t, 1.0, result.ConfidenceScore)
}

func TestDraftGenerator_Success(t *testing.T) {
	adapter := &mockAdapter{
		outputStr: `{
			"customer_update_draft": "Valid draft directly from AI.",
			"confidence_score": 0.95
		}`,
	}
	generator := NewDraftGenerator(adapter, DraftGeneratorConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	input := dto.CustomerUpdateDraftInput{
		OrderID:       1006,
		ExceptionType: "STUCK_ORDER",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err)
	assert.False(t, result.FallbackUsed)
	assert.Equal(t, "Valid draft directly from AI.", result.CustomerUpdateDraft)
	assert.Equal(t, 0.95, result.ConfidenceScore)
}
