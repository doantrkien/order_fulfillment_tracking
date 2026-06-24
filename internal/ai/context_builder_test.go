package ai

import (
	"testing"

	"main/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestProcessDriverNote_MaskPhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal Phone", "Gọi 0987654321 nhé", "Gọi [REDACTED_PHONE] nhé"},
		{"Phone with country code", "SĐT +84987654321 hoặc 84987654321", "SĐT [REDACTED_PHONE] hoặc [REDACTED_PHONE]"},
		{"Phone with dashes", "Liên hệ 0123-456-789", "Liên hệ [REDACTED_PHONE]"},
		{"Phone with spaces", "Số 012 345 6789", "Số [REDACTED_PHONE]"},
		{"No phone", "Giao hàng cẩn thận", "Giao hàng cẩn thận"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ProcessDriverNote(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestProcessDriverNote_Truncate(t *testing.T) {
	// Create a string with 1005 runes
	input := ""
	for i := 0; i < 1005; i++ {
		input += "á" // Vietnamese char, 2 bytes
	}

	actual := ProcessDriverNote(input)
	
	// Should be truncated to 1000 runes
	runes := []rune(actual)
	assert.Equal(t, 1000, len(runes))
}

func TestProcessDriverNote_Sanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Control chars removed", "Line 1\nLine 2\tTabbed\x00Hidden", "Line 1 Line 2 TabbedHidden"},
		{"Quotes replaced", `He said "hello"`, "He said 'hello'"},
		{"Extra spaces removed", "Too   many    spaces", "Too many spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ProcessDriverNote(tt.input)
			// ProcessDriverNote removes newlines and tabs due to strings.Fields, which is expected for a clean note
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestSanitizeEvents(t *testing.T) {
	events := make([]models.AIEvent, 15)
	for i := 0; i < 15; i++ {
		events[i] = models.AIEvent{UpdatedBy: string(rune('A' + i))}
	}

	actual := SanitizeEvents(events)
	assert.Equal(t, 10, len(actual))
	// Should keep the last 10
	assert.Equal(t, "F", actual[0].UpdatedBy) // 15 - 10 = 5 (index 5 is F)
	assert.Equal(t, "O", actual[9].UpdatedBy) // index 14 is O
}

func TestPrepareContextData(t *testing.T) {
	ctx := &models.AIContext{
		OrderID:       123,
		CurrentStatus: "shipped",
		TotalAmount:   500000,
		Events: []models.AIEvent{
			{PreviousStatus: "created", NewStatus: "paid"},
		},
	}

	jsonStr, err := PrepareContextData(ctx)
	assert.NoError(t, err)

	// Check if key fields are in JSON
	assert.Contains(t, jsonStr, `"order_id":123`)
	assert.Contains(t, jsonStr, `"current_status":"shipped"`)
	assert.Contains(t, jsonStr, `[REDACTED_CUSTOMER_NAME]`)
	assert.Contains(t, jsonStr, `"events":[{`)
}

func TestBuildDynamicPrompt(t *testing.T) {
	// Need to check if the prompt builder includes our mock vars
	prompt := BuildDynamicPrompt("exception", `{"test":"context"}`, "[KNOWLEDGE BASE]")
	assert.Contains(t, prompt, `{"test":"context"}`)
	assert.Contains(t, prompt, "[KNOWLEDGE BASE]")
}