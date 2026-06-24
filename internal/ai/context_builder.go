package ai

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"main/internal/models"
)

// Regex để tìm số điện thoại Việt Nam (vd: 0987654321, 84987654321, +84987654321, 0123.456.789, 0123 456 789, 0123-456-789)
var phoneRegex = regexp.MustCompile(`(?:\+84|84|0)[0-9]{1,3}[\s\.\-]?[0-9]{3}[\s\.\-]?[0-9]{3,4}`)

// ProcessDriverNote thực hiện làm sạch driver note theo đúng thứ tự 4 bước:
// 1. Mask PII (Che số điện thoại)
// 2. Truncate (Cắt max 1000 ký tự theo rune)
// 3. Sanitize (Loại bỏ ký tự điều khiển, chuẩn hóa khoảng trắng)
func ProcessDriverNote(note string) string {
	if note == "" {
		return ""
	}

	// 1. Mask PII: Thay thế số điện thoại bằng [REDACTED_PHONE]
	masked := phoneRegex.ReplaceAllString(note, "[REDACTED_PHONE]")

	// 2. Truncate by rune: Max 1000 ký tự, giữ nguyên các ký tự tiếng Việt
	runes := []rune(masked)
	if len(runes) > 1000 {
		runes = runes[:1000]
	}
	truncated := string(runes)

	// 3. Sanitize: Chống Prompt Injection
	// Loại bỏ ký tự điều khiển (giữ lại \n và \t)
	var sb strings.Builder
	for _, r := range truncated {
		if (r >= 0 && r < 32 && r != '\n' && r != '\t') || r == 127 {
			continue // Skip control characters
		}
		// Replace " with '
		if r == '"' {
			sb.WriteRune('\'')
			continue
		}
		sb.WriteRune(r)
	}
	
	// Chuẩn hóa khoảng trắng: thay nhiều khoảng trắng liên tiếp bằng 1 khoảng trắng
	sanitized := strings.Join(strings.Fields(sb.String()), " ")
	return sanitized
}

// SanitizeEvents giữ lại tối đa 10 sự kiện gần nhất
func SanitizeEvents(events []models.AIEvent) []models.AIEvent {
	if len(events) <= 10 {
		return events
	}
	return events[len(events)-10:]
}

// PrepareContextData chuẩn bị dữ liệu (thực hiện Marshal JSON sau khi đã làm sạch)
func PrepareContextData(aiCtx *models.AIContext) (string, error) {
	// Tạo bản sao để tránh thay đổi struct gốc
	cleanEvents := SanitizeEvents(aiCtx.Events)
	
	dataMap := map[string]interface{}{
		"order_id":         aiCtx.OrderID,
		"current_status":   aiCtx.CurrentStatus,
		"total_amount":     aiCtx.TotalAmount,
		"created_at":       aiCtx.CreatedAt,
		// Tên và địa chỉ đã được che ở nơi khác nếu cần, hoặc có thể che luôn tại đây
		"customer_name":    "[REDACTED_CUSTOMER_NAME]",
		"shipping_address": "[REDACTED_SHIPPING_ADDRESS]",
		"events":           cleanEvents,
	}

	// 4. JSON Marshal
	bytes, err := json.Marshal(dataMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal context data: %w", err)
	}

	return string(bytes), nil
}

// BuildDynamicPrompt lắp ráp system instruction, task request, kbSnippet và context json.
func BuildDynamicPrompt(taskType string, sanitizedContext string, kbSnippet string) string {
	var promptBuilder strings.Builder
	
	// 1. System Instruction
	promptBuilder.WriteString(systemInstructionMD)
	promptBuilder.WriteString("\n\n")

	// 2. Task Markdown & Inject
	var taskTemplate string
	if taskType == "draft" {
		taskTemplate = taskDraftMD
	} else {
		taskTemplate = taskExceptionMD
	}

	// Nếu kbSnippet rỗng, thì thay {{KNOWLEDGE_BASE_INJECTION}} thành rỗng
	taskTemplate = strings.ReplaceAll(taskTemplate, "{{KNOWLEDGE_BASE_INJECTION}}", kbSnippet)
	
	// Các taskTemplate cũ sử dụng nhiều {{VARIABLE}}, tuy nhiên ta có thể
	// append {{CONTEXT_JSON}} nếu nó chưa được define trong template để tương thích
	taskTemplate = strings.ReplaceAll(taskTemplate, "{{CONTEXT_JSON}}", sanitizedContext)

	// Inject json context string (vì template cũ có thể không có {{CONTEXT_JSON}})
	if !strings.Contains(taskTemplate, sanitizedContext) {
		taskTemplate += "\n\n[CONTEXT JSON]\n" + sanitizedContext
	}

	promptBuilder.WriteString(taskTemplate)

	return promptBuilder.String()
}
