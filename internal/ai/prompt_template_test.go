package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizePromptContext_EventTimelineCapped(t *testing.T) {
	tests := []struct {
		name       string
		numEvents  int
		wantEvents int
	}{
		{"under_limit_10", 10, 10},
		{"at_limit_50", 50, 50},
		{"over_limit_100", 100, 50},
		{"way_over_limit_500", 500, 50},
		{"zero_events", 0, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			events := make([]EventTimelineEntry, tc.numEvents)
			for i := range events {
				events[i] = EventTimelineEntry{
					FromStatus: "created",
					ToStatus:   "paid",
					UpdatedBy:  "admin_1",
					EventAt:    "2026-06-01 10:00:00",
				}
			}

			ctx := &ExceptionPromptContext{
				OrderID:       1,
				EventTimeline: events,
			}

			SanitizePromptContext(ctx)
			assert.Equal(t, tc.wantEvents, len(ctx.EventTimeline))
		})
	}
}

func TestSanitizePromptContext_KeepsMostRecentEvents(t *testing.T) {
	events := make([]EventTimelineEntry, 60)
	for i := range events {
		events[i] = EventTimelineEntry{
			FromStatus: "created",
			ToStatus:   "paid",
			UpdatedBy:  "admin_1",
			EventAt:    strings.Repeat("x", i), // use as unique marker
		}
	}

	ctx := &ExceptionPromptContext{
		OrderID:       1,
		EventTimeline: events,
	}

	SanitizePromptContext(ctx)

	assert.Equal(t, 50, len(ctx.EventTimeline))
	assert.Equal(t, events[10].EventAt, ctx.EventTimeline[0].EventAt)
	assert.Equal(t, events[59].EventAt, ctx.EventTimeline[49].EventAt)
}

func TestSanitizePromptContext_DriverNotesCapped(t *testing.T) {
	tests := []struct {
		name       string
		notesLen   int
		wantCapped bool
	}{
		{"under_limit_100", 100, false},
		{"at_limit_500", 500, false},
		{"over_limit_501", 501, true},
		{"way_over_600", 600, true},
		{"empty_notes", 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			notes := strings.Repeat("n", tc.notesLen)
			ctx := &ExceptionPromptContext{
				OrderID:     1,
				DriverNotes: notes,
			}

			SanitizePromptContext(ctx)

			if tc.wantCapped {
				assert.LessOrEqual(t, len(ctx.DriverNotes), MaxDriverNotesLength+len("...[truncated]"))
				assert.True(t, strings.HasSuffix(ctx.DriverNotes, "...[truncated]"))
			} else {
				assert.Equal(t, tc.notesLen, len(ctx.DriverNotes))
			}
		})
	}
}

func TestSanitizePromptContext_PIIRedacted(t *testing.T) {
	ctx := &ExceptionPromptContext{
		OrderID:         1,
		CustomerName:    "Nguyen Van A",
		ShippingAddress: "123 Le Loi, HCM",
	}

	SanitizePromptContext(ctx)

	assert.Equal(t, "[REDACTED_CUSTOMER_NAME]", ctx.CustomerName)
	assert.Equal(t, "[REDACTED_SHIPPING_ADDRESS]", ctx.ShippingAddress)
}

func TestSanitizePromptContext_EmptyPIINotRedacted(t *testing.T) {
	ctx := &ExceptionPromptContext{
		OrderID:         1,
		CustomerName:    "",
		ShippingAddress: "",
	}

	SanitizePromptContext(ctx)

	assert.Equal(t, "", ctx.CustomerName)
	assert.Equal(t, "", ctx.ShippingAddress)
}

func TestBuildExceptionAnalysisPrompt_ContainsAllSections(t *testing.T) {
	ctx := ExceptionPromptContext{
		OrderID:         123,
		CurrentStatus:   "shipped",
		TotalAmount:     500000,
		CustomerName:    "Nguyen Van A",
		ShippingAddress: "123 HCM",
		CreatedAt:       "2026-06-01 08:00:00",
		EventTimeline: []EventTimelineEntry{
			{FromStatus: "created", ToStatus: "paid", UpdatedBy: "admin_1", EventAt: "2026-06-01 09:00:00"},
			{FromStatus: "paid", ToStatus: "packed", UpdatedBy: "admin_1", EventAt: "2026-06-01 10:00:00"},
		},
		DriverNotes: "Package looks damaged",
	}

	SanitizePromptContext(&ctx)
	knowledge := ClassifyDriverNote(ctx.DriverNotes)
	prompt := BuildExceptionAnalysisPrompt(ctx, knowledge)

	assert.Contains(t, prompt, "[SYSTEM]")
	assert.Contains(t, prompt, "[CONTEXT]")
	assert.Contains(t, prompt, "[KNOWLEDGE BASE]")
	assert.Contains(t, prompt, "[TASK]")
	assert.Contains(t, prompt, "[OUTPUT FORMAT]")
	assert.Contains(t, prompt, "[CONSTRAINTS]")

	assert.Contains(t, prompt, "Order ID: 123")
	assert.Contains(t, prompt, "Current Status: shipped")
	assert.Contains(t, prompt, "500000 VND")
	// PII should be redacted in the prompt — real names must NOT appear
	assert.NotContains(t, prompt, "Nguyen Van A")
	assert.NotContains(t, prompt, "123 HCM")
	assert.Contains(t, prompt, "[REDACTED_CUSTOMER_NAME]")
	assert.Contains(t, prompt, "[REDACTED_SHIPPING_ADDRESS]")
	assert.Contains(t, prompt, "Package looks damaged")

	assert.Contains(t, prompt, "created → paid")
	assert.Contains(t, prompt, "paid → packed")

	assert.Contains(t, prompt, "MUST NOT suggest or imply any automatic order status changes")
	assert.Contains(t, prompt, "MUST NOT suggest or trigger any refund")
	assert.Contains(t, prompt, "MUST NOT suggest sending any messages")
}

func TestBuildExceptionAnalysisPrompt_EmptyTimeline(t *testing.T) {
	ctx := ExceptionPromptContext{
		OrderID:       1,
		CurrentStatus: "created",
	}

	prompt := BuildExceptionAnalysisPrompt(ctx, nil)
	assert.Contains(t, prompt, "(no events recorded)")
}

func TestBuildExceptionAnalysisPrompt_NoDriverNotes(t *testing.T) {
	ctx := ExceptionPromptContext{
		OrderID:       1,
		CurrentStatus: "created",
		DriverNotes:   "",
	}

	prompt := BuildExceptionAnalysisPrompt(ctx, nil)
	assert.NotContains(t, prompt, "Driver Note:")
}

// ── ClassifyDriverNote tests ─────────────────────────────────────────────────

func TestClassifyDriverNote_AlwaysIncludesStateMachine(t *testing.T) {
	entries := ClassifyDriverNote("")
	ids := kbIDs(entries)
	assert.Contains(t, ids, "state_machine")
}

func TestClassifyDriverNote_DeliveryKeywords(t *testing.T) {
	cases := []string{
		"customer not home",
		"xe hỏng trên đường",
		"accident on highway",
		"package lost",
		"could not deliver",
		"giao thất bại",
		"bad weather",
	}
	for _, note := range cases {
		entries := ClassifyDriverNote(note)
		ids := kbIDs(entries)
		assert.Contains(t, ids, "delivery_failure", "note: %q", note)
	}
}

func TestClassifyDriverNote_CancellationKeywords(t *testing.T) {
	entries := ClassifyDriverNote("khách hàng muốn hủy đơn hàng")
	ids := kbIDs(entries)
	assert.Contains(t, ids, "cancellation_anomaly")
}

func TestClassifyDriverNote_RefundKeywords(t *testing.T) {
	entries := ClassifyDriverNote("khách yêu cầu hoàn tiền")
	ids := kbIDs(entries)
	assert.Contains(t, ids, "refund_anomaly")
}

func TestClassifyDriverNote_StuckKeywords(t *testing.T) {
	entries := ClassifyDriverNote("order is delayed, no progress for 3 days")
	ids := kbIDs(entries)
	assert.Contains(t, ids, "stuck_order")
}

func TestClassifyDriverNote_UnknownNote_FallsBackToDeliveryAndDataIntegrity(t *testing.T) {
	// A non-empty note that matches no specific keyword should still get
	// delivery_failure and data_integrity so the AI has useful context.
	entries := ClassifyDriverNote("driver arrived at destination")
	ids := kbIDs(entries)
	assert.Contains(t, ids, "state_machine")
	assert.Contains(t, ids, "delivery_failure")
	assert.Contains(t, ids, "data_integrity")
}

func TestBuildExceptionAnalysisPrompt_InjectsOnlyRelevantKB(t *testing.T) {
	ctx := ExceptionPromptContext{
		OrderID:       999,
		CurrentStatus: "shipped",
		DriverNotes:   "vehicle breakdown on the way",
	}
	knowledge := ClassifyDriverNote(ctx.DriverNotes)
	prompt := BuildExceptionAnalysisPrompt(ctx, knowledge)

	// Delivery failure KB should be present
	assert.Contains(t, prompt, "Delivery Failure")
	// State machine KB should always be present
	assert.Contains(t, prompt, "Order State Machine")
	// Cancellation and refund KB should NOT be present (not in note)
	assert.NotContains(t, prompt, "Cancellation Anomaly")
	assert.NotContains(t, prompt, "Refund Anomaly")
}

// kbIDs extracts the ID field from a slice of KnowledgeEntry for easy assertion.
func kbIDs(entries []KnowledgeEntry) []string {
	ids := make([]string, len(entries))
	for i, e := range entries {
		ids[i] = e.ID
	}
	return ids
}
