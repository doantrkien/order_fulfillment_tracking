package ai

import (
	"context"
	"errors"
	"main/constant"
	"main/internal/ai"
	dto_ai "main/internal/dto/ai"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDraftGenerator_AIDisabled(t *testing.T) {
	generator := ai.NewDraftGenerator(nil, ai.DraftGeneratorConfig{
		AIEnabled: false,
		AITimeout: 10 * time.Second,
	})

	input := dto_ai.CustomerUpdateDraftInput{
		OrderID:       1001,
		ExceptionType: "STUCK_ORDER",
		Tone:          "neutral",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, constant.FallbackReasonDisabled, result.FallbackReason)
	assert.Contains(t, result.CustomerUpdateDraft, "slower than expected")
	assert.Equal(t, 1.0, result.ConfidenceScore)
}

func TestDraftGenerator_AIReturnsError(t *testing.T) {
	adapter := &mockAdapter{
		err: errors.New("connection reset by peer"),
	}
	generator := ai.NewDraftGenerator(adapter, ai.DraftGeneratorConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	input := dto_ai.CustomerUpdateDraftInput{
		OrderID:       1002,
		ExceptionType: "DELIVERY_FAILURE",
		Tone:          "apologetic",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err) // DraftGenerator should never return error on AI failure
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, constant.FallbackReasonConnectionError, result.FallbackReason)
	assert.Contains(t, result.CustomerUpdateDraft, "order to .")
	assert.Equal(t, 1.0, result.ConfidenceScore)
}

func TestDraftGenerator_AIReturnsTimeout(t *testing.T) {
	adapter := &mockAdapter{
		err: context.DeadlineExceeded,
	}
	generator := ai.NewDraftGenerator(adapter, ai.DraftGeneratorConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	input := dto_ai.CustomerUpdateDraftInput{
		OrderID:       1003,
		ExceptionType: "DELIVERY_FAILURE",
		Tone:          "apologetic",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, constant.FallbackReasonTimeout, result.FallbackReason)
	assert.Contains(t, result.CustomerUpdateDraft, "issue occurred during the shipment")
	assert.Equal(t, 1.0, result.ConfidenceScore)
}

func TestDraftGenerator_AIReturnsInvalidResponse(t *testing.T) {
	adapter := &mockAdapter{
		outputStr: "{invalid-json}",
	}
	generator := ai.NewDraftGenerator(adapter, ai.DraftGeneratorConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	input := dto_ai.CustomerUpdateDraftInput{
		OrderID:       1004,
		ExceptionType: "OTHER",
		Tone:          "apologetic",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, constant.FallbackReasonInvalidResponse, result.FallbackReason)
	assert.Contains(t, result.CustomerUpdateDraft, "issue that has arisen regarding your order")
	assert.Equal(t, 1.0, result.ConfidenceScore)
}

func TestDraftGenerator_AIReturnsLowConfidence(t *testing.T) {
	adapter := &mockAdapter{
		outputStr: `{
			"customer_update_draft": "Apology message...",
			"confidence_score": 0.3
		}`,
	}
	generator := ai.NewDraftGenerator(adapter, ai.DraftGeneratorConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	input := dto_ai.CustomerUpdateDraftInput{
		OrderID:       1005,
		ExceptionType: "DELIVERY_FAILURE",
		Tone:          "apologetic",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.FallbackUsed)
	assert.Equal(t, constant.FallbackReasonLowConfidence, result.FallbackReason)
	assert.Contains(t, result.CustomerUpdateDraft, "issue occurred during the shipment")
	assert.Equal(t, 1.0, result.ConfidenceScore)
}

func TestDraftGenerator_Success(t *testing.T) {
	adapter := &mockAdapter{
		outputStr: `{
			"customer_update_draft": "Valid draft directly from AI.",
			"confidence_score": 0.95
		}`,
	}
	generator := ai.NewDraftGenerator(adapter, ai.DraftGeneratorConfig{
		AIEnabled: true,
		AITimeout: 10 * time.Second,
	})

	input := dto_ai.CustomerUpdateDraftInput{
		OrderID:       1006,
		ExceptionType: "OTHER",
	}

	result, err := generator.Generate(context.Background(), input)

	require.NoError(t, err)
	assert.False(t, result.FallbackUsed)
	assert.Equal(t, "Valid draft directly from AI.", result.CustomerUpdateDraft)
	assert.Equal(t, 0.95, result.ConfidenceScore)
}
