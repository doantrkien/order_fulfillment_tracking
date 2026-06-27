package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"main/internal/dto"
)

type DraftResult struct {
	CustomerUpdateDraft string
	ConfidenceScore     float64
	FallbackUsed        bool
	FallbackReason      string
	DurationMs          int
	RawResponse         string
}

type DraftGeneratorConfig struct {
	AIEnabled bool
	AITimeout time.Duration
}

type DraftGenerator struct {
	adapter AIAdapter
	config  DraftGeneratorConfig
}

func NewDraftGenerator(adapter AIAdapter, config DraftGeneratorConfig) *DraftGenerator {
	return &DraftGenerator{
		adapter: adapter,
		config:  config,
	}
}

func (dg *DraftGenerator) Generate(ctx context.Context, input dto.CustomerUpdateDraftInput) (*DraftResult, error) {

	input.BaselineDraft = buildFallbackDraftMessage(input)

	if !dg.config.AIEnabled || !shouldCallAI(input) {
		reason := FallbackReasonTemplateSufficient
		if !dg.config.AIEnabled {
			reason = FallbackReasonDisabled
		}
		return &DraftResult{
			CustomerUpdateDraft: input.BaselineDraft,
			ConfidenceScore:     1.0,
			FallbackUsed:        true,
			FallbackReason:      reason,
			DurationMs:          0,
		}, nil
	}

	start := time.Now()
	timeoutCtx, cancel := context.WithTimeout(ctx, dg.config.AITimeout)
	defer cancel()

	rawText, err := dg.adapter.DraftCustomerUpdate(timeoutCtx, input)
	durationMs := int(time.Since(start).Milliseconds())
	if err != nil {
		reason := ClassifyError(err)
		return dg.fallback(input, reason, durationMs, ""), nil
	}

	cleaned := stripDraftFences(rawText)
	var output dto.CustomerUpdateDraftOutput
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return dg.fallback(input, FallbackReasonInvalidResponse, durationMs, rawText), nil
	}

	if strings.TrimSpace(output.CustomerUpdateDraft) == "" {
		return dg.fallback(input, FallbackReasonInvalidResponse, durationMs, rawText), nil
	}

	if output.ConfidenceScore < ConfidenceThreshold {
		return dg.fallback(input, FallbackReasonLowConfidence, durationMs, rawText), nil
	}

	return &DraftResult{
		CustomerUpdateDraft: output.CustomerUpdateDraft,
		ConfidenceScore:     output.ConfidenceScore,
		FallbackUsed:        false,
		DurationMs:          durationMs,
		RawResponse:         rawText,
	}, nil
}

// fallback generates a safe template-based draft message when AI is unavailable.
func (dg *DraftGenerator) fallback(input dto.CustomerUpdateDraftInput, reason string, durationMs int, rawResponse string) *DraftResult {
	message := input.BaselineDraft
	if message == "" {
		message = buildFallbackDraftMessage(input)
	}
	return &DraftResult{
		CustomerUpdateDraft: message,
		ConfidenceScore:     1.0,
		FallbackUsed:        true,
		FallbackReason:      reason,
		DurationMs:          durationMs,
		RawResponse:         rawResponse,
	}
}

// isInternalSystemError trả về true cho các loại lỗi hệ thống nội bộ.
// Các loại lỗi này không cần AI diễn giải thêm — Template tĩnh đã đủ an toàn và trung lập.
func isInternalSystemError(exceptionType string) bool {
	switch exceptionType {
	case "INVALID_TRANSITION", "SKIPPED_STATUS", "DUPLICATE_EVENT":
		return true
	}
	return false
}

// shouldCallAI quyết định có cần gọi AI hay dùng Template tĩnh.
//
// Ma trận quyết định:
//   - OTHER → LUÔN gọi AI (LikelyReason là text tự do, Template không diễn giải được)
//   - INVALID_TRANSITION / SKIPPED_STATUS / DUPLICATE_EVENT → KHÔNG gọi AI (Template đủ, an toàn)
//   - ALTERNATIVE_DELIVERY → KHÔNG gọi AI (Giao thành công, Template FYI đủ)
//   - STUCK_ORDER / DELIVERY_FAILURE + neutral/informative + email → Template
//   - STUCK_ORDER / DELIVERY_FAILURE + apologetic/proactive → AI (cần giọng điệu)
//   - Bất kỳ exception nào + channel=sms → AI (cần rút ngắn)
func shouldCallAI(input dto.CustomerUpdateDraftInput) bool {
	channel := strings.ToLower(strings.TrimSpace(input.Channel))
	tone := strings.ToLower(strings.TrimSpace(input.Tone))
	var needAI bool

	switch {
	case input.ExceptionType == "OTHER":
		// Template không thể diễn giải LikelyReason tự do từ exception analysis
		needAI = true

	case isInternalSystemError(input.ExceptionType):
		// Lỗi kỹ thuật nội bộ — không để AI "sáng tác" thêm cho khách hàng
		needAI = false

	case input.ExceptionType == "ALTERNATIVE_DELIVERY":
		// Giao thay thế thành công — Template FYI đủ an toàn cho khách hàng
		needAI = false

	case channel == "sms":
		// SMS cần rút ngắn mạnh, Template tĩnh quá dài
		needAI = true

	case tone != "" && tone != "neutral" && tone != "informative":
		// Tone đặc biệt (apologetic, proactive) cần AI viết lại giọng điệu
		needAI = true

	default:
		// Còn lại: lỗi vận hành (STUCK_ORDER/DELIVERY_FAILURE) + tone cơ bản + email
		needAI = false
	}

	fmt.Printf("[DEBUG][shouldCallAI] ExceptionType: %s | Tone: %q | Channel: %q | NeedAI: %v\n",
		input.ExceptionType, tone, channel, needAI)

	return needAI
}

func buildFallbackDraftMessage(input dto.CustomerUpdateDraftInput) string {
	return GetFallbackTemplate(input.ExceptionType, input.CustomerName, input.ShippingAddress, input.CurrentStatus)
}

func stripDraftFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	return s
}
