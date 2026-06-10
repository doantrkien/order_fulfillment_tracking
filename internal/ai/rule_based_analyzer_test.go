package ai

import (
	"main/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDetectInvalidTransitions(t *testing.T) {
	tests := []struct {
		name     string
		events   []models.AIEvent
		wantType string
		wantNil  bool
	}{
		{
			name: "valid transition created→paid",
			events: []models.AIEvent{
				{PreviousStatus: models.ORDER_STATUS_CREATED, NewStatus: models.ORDER_STATUS_PAID},
			},
			wantNil: true,
		},
		{
			name: "invalid transition created→delivered",
			events: []models.AIEvent{
				{PreviousStatus: models.ORDER_STATUS_CREATED, NewStatus: models.ORDER_STATUS_DELIVERED},
			},
			wantType: "INVALID_TRANSITION",
		},
		{
			name: "invalid transition packed→cancelled",
			events: []models.AIEvent{
				{PreviousStatus: models.ORDER_STATUS_CREATED, NewStatus: models.ORDER_STATUS_PAID},
				{PreviousStatus: models.ORDER_STATUS_PAID, NewStatus: models.ORDER_STATUS_PACKED},
				{PreviousStatus: models.ORDER_STATUS_PACKED, NewStatus: models.ORDER_STATUS_CANCELLED},
			},
			wantType: "INVALID_TRANSITION",
		},
		{
			name:    "empty events",
			events:  []models.AIEvent{},
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectInvalidTransitions(tc.events)
			if tc.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.wantType, result.ExceptionType)
				assert.Equal(t, "CRITICAL", result.Severity)
				assert.Equal(t, 1.0, result.ConfidenceScore)
			}
		})
	}
}

func TestDetectDuplicateEvents(t *testing.T) {
	tests := []struct {
		name     string
		events   []models.AIEvent
		wantType string
		wantNil  bool
	}{
		{
			name: "no duplicates",
			events: []models.AIEvent{
				{NewStatus: models.ORDER_STATUS_PAID},
				{NewStatus: models.ORDER_STATUS_PACKED},
			},
			wantNil: true,
		},
		{
			name: "consecutive duplicate",
			events: []models.AIEvent{
				{NewStatus: models.ORDER_STATUS_PAID},
				{NewStatus: models.ORDER_STATUS_PAID},
			},
			wantType: "DUPLICATE_EVENT",
		},
		{
			name: "single event - no duplicate possible",
			events: []models.AIEvent{
				{NewStatus: models.ORDER_STATUS_PAID},
			},
			wantNil: true,
		},
		{
			name:    "empty events",
			events:  []models.AIEvent{},
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectDuplicateEvents(tc.events)
			if tc.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.wantType, result.ExceptionType)
				assert.Equal(t, "LOW", result.Severity)
				assert.Equal(t, 1.0, result.ConfidenceScore)
			}
		})
	}
}

func TestDetectSkippedStatuses(t *testing.T) {
	tests := []struct {
		name     string
		events   []models.AIEvent
		wantType string
		wantNil  bool
	}{
		{
			name: "normal progression - no skip",
			events: []models.AIEvent{
				{NewStatus: models.ORDER_STATUS_PAID},
				{NewStatus: models.ORDER_STATUS_PACKED},
				{NewStatus: models.ORDER_STATUS_SHIPPED},
			},
			wantNil: true,
		},
		{
			name: "skipped packed status",
			events: []models.AIEvent{
				{NewStatus: models.ORDER_STATUS_PAID},
				{NewStatus: models.ORDER_STATUS_SHIPPED},
			},
			wantType: "SKIPPED_STATUS",
		},
		{
			name: "skipped paid and packed",
			events: []models.AIEvent{
				{NewStatus: models.ORDER_STATUS_SHIPPED},
			},
			wantType: "SKIPPED_STATUS",
		},
		{
			name: "only paid - no skip",
			events: []models.AIEvent{
				{NewStatus: models.ORDER_STATUS_PAID},
			},
			wantNil: true,
		},
		{
			name:    "empty events",
			events:  []models.AIEvent{},
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectSkippedStatuses(tc.events)
			if tc.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.wantType, result.ExceptionType)
				assert.Equal(t, "HIGH", result.Severity)
			}
		})
	}
}

func TestDetectStuckOrder(t *testing.T) {
	baseTime := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		aiCtx        *models.AIContext
		now          time.Time
		wantType     string
		wantSeverity string
		wantNil      bool
	}{
		{
			name: "created status - not stuck (under 24h)",
			aiCtx: &models.AIContext{
				CurrentStatus: models.ORDER_STATUS_CREATED,
				CreatedAt:     baseTime,
			},
			now:     baseTime.Add(23 * time.Hour),
			wantNil: true,
		},
		{
			name: "created status - stuck (over 24h)",
			aiCtx: &models.AIContext{
				CurrentStatus: models.ORDER_STATUS_CREATED,
				CreatedAt:     baseTime,
			},
			now:          baseTime.Add(25 * time.Hour),
			wantType:     "STUCK_ORDER",
			wantSeverity: "MEDIUM",
		},
		{
			name: "created status - escalated (over 48h = 2×24h)",
			aiCtx: &models.AIContext{
				CurrentStatus: models.ORDER_STATUS_CREATED,
				CreatedAt:     baseTime,
			},
			now:          baseTime.Add(49 * time.Hour),
			wantType:     "STUCK_ORDER",
			wantSeverity: "CRITICAL",
		},
		{
			name: "shipped status - stuck (over 72h)",
			aiCtx: &models.AIContext{
				CurrentStatus: models.ORDER_STATUS_SHIPPED,
				CreatedAt:     baseTime,
				Events: []models.AIEvent{
					{EventAt: baseTime.Add(1 * time.Hour), NewStatus: models.ORDER_STATUS_SHIPPED},
				},
			},
			now:          baseTime.Add(74 * time.Hour),
			wantType:     "STUCK_ORDER",
			wantSeverity: "HIGH",
		},
		{
			name: "delivered (terminal) - never stuck",
			aiCtx: &models.AIContext{
				CurrentStatus: models.ORDER_STATUS_DELIVERED,
				CreatedAt:     baseTime,
			},
			now:     baseTime.Add(200 * time.Hour),
			wantNil: true,
		},
		{
			name: "cancelled (terminal) - never stuck",
			aiCtx: &models.AIContext{
				CurrentStatus: models.ORDER_STATUS_CANCELLED,
				CreatedAt:     baseTime,
			},
			now:     baseTime.Add(200 * time.Hour),
			wantNil: true,
		},
		{
			name: "uses last event time not order creation",
			aiCtx: &models.AIContext{
				CurrentStatus: models.ORDER_STATUS_PAID,
				CreatedAt:     baseTime,
				Events: []models.AIEvent{
					{EventAt: baseTime.Add(47 * time.Hour), NewStatus: models.ORDER_STATUS_PAID},
				},
			},
			// 47h after creation, event happened. Now is 50h after creation.
			// Age from last event = 3h, which is < 48h threshold → not stuck.
			now:     baseTime.Add(50 * time.Hour),
			wantNil: true,
		},
		{
			name: "boundary - exactly at threshold",
			aiCtx: &models.AIContext{
				CurrentStatus: models.ORDER_STATUS_CREATED,
				CreatedAt:     baseTime,
			},
			now:          baseTime.Add(24 * time.Hour), // exactly at threshold
			wantType:     "STUCK_ORDER",
			wantSeverity: "MEDIUM",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectStuckOrder(tc.aiCtx, tc.now)
			if tc.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.wantType, result.ExceptionType)
				assert.Equal(t, tc.wantSeverity, result.Severity)
				assert.Equal(t, 1.0, result.ConfidenceScore)
			}
		})
	}
}

func TestAnalyzeByRules_PriorityOrder(t *testing.T) {
	now := time.Now()

	t.Run("invalid transition takes priority over duplicate", func(t *testing.T) {
		aiCtx := &models.AIContext{
			CurrentStatus: models.ORDER_STATUS_SHIPPED,
			CreatedAt:     now.Add(-100 * time.Hour),
			Events: []models.AIEvent{
				{PreviousStatus: models.ORDER_STATUS_CREATED, NewStatus: models.ORDER_STATUS_PAID, EventAt: now.Add(-99 * time.Hour)},
				// Invalid transition
				{PreviousStatus: models.ORDER_STATUS_PAID, NewStatus: models.ORDER_STATUS_DELIVERED, EventAt: now.Add(-98 * time.Hour)},
				// Also a duplicate
				{PreviousStatus: models.ORDER_STATUS_PAID, NewStatus: models.ORDER_STATUS_DELIVERED, EventAt: now.Add(-97 * time.Hour)},
			},
		}
		result := AnalyzeByRules(aiCtx, now)
		assert.NotNil(t, result)
		assert.Equal(t, "INVALID_TRANSITION", result.ExceptionType)
	})

	t.Run("duplicate takes priority over skipped status", func(t *testing.T) {
		aiCtx := &models.AIContext{
			CurrentStatus: models.ORDER_STATUS_SHIPPED,
			CreatedAt:     now.Add(-10 * time.Hour),
			Events: []models.AIEvent{
				{PreviousStatus: models.ORDER_STATUS_CREATED, NewStatus: models.ORDER_STATUS_PAID, EventAt: now.Add(-9 * time.Hour)},
				// Valid transition, then duplicate of the same status
				{PreviousStatus: models.ORDER_STATUS_PAID, NewStatus: models.ORDER_STATUS_PACKED, EventAt: now.Add(-8 * time.Hour)},
				{PreviousStatus: models.ORDER_STATUS_PAID, NewStatus: models.ORDER_STATUS_PACKED, EventAt: now.Add(-7 * time.Hour)},
				// Skip shipped, go straight to delivered via shipped (would be SKIPPED_STATUS if not for duplicate)
			},
		}
		result := AnalyzeByRules(aiCtx, now)
		assert.NotNil(t, result)
		assert.Equal(t, "DUPLICATE_EVENT", result.ExceptionType)
	})

	t.Run("no exception detected returns nil", func(t *testing.T) {
		aiCtx := &models.AIContext{
			CurrentStatus: models.ORDER_STATUS_PAID,
			CreatedAt:     now.Add(-1 * time.Hour),
			Events: []models.AIEvent{
				{PreviousStatus: models.ORDER_STATUS_CREATED, NewStatus: models.ORDER_STATUS_PAID, EventAt: now.Add(-30 * time.Minute)},
			},
		}
		result := AnalyzeByRules(aiCtx, now)
		assert.Nil(t, result)
	})
}
