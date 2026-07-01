package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"main/internal/dto"
)

const (
	ExceptionAlternativeDelivery = "ALTERNATIVE_DELIVERY"
	ExceptionOther               = "OTHER"

	ChannelSMS     = "sms"
	ToneApologetic = "apologetic"
	ToneProactive  = "proactive"
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

	// Scoring parameters
	ScoreChannelSMS              int
	ScoreToneApologeticProactive int
	ScoreLongReason              int
	LikelyReasonLengthThreshold  int
	AIScoreThreshold             int
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

	if !dg.config.AIEnabled || !dg.shouldCallAI(input) {
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
func (dg *DraftGenerator) shouldCallAI(input dto.CustomerUpdateDraftInput) bool {
	channel := strings.ToLower(strings.TrimSpace(input.Channel))
	tone := strings.ToLower(strings.TrimSpace(input.Tone))

	// 1. HARD RULES (Veto - Phủ quyết tuyệt đối)
	if isInternalSystemError(input.ExceptionType) {
		fmt.Printf("[DEBUG][shouldCallAI] VETO: Internal System Error (%s)\n", input.ExceptionType)
		return false
	}
	if input.ExceptionType == ExceptionAlternativeDelivery {
		fmt.Printf("[DEBUG][shouldCallAI] VETO: Alternative Delivery\n")
		return false
	}

	// 2. MUST-HAVE RULES (Bắt buộc dùng AI)
	if input.ExceptionType == ExceptionOther {
		fmt.Printf("[DEBUG][shouldCallAI] MUST: Exception Type is OTHER\n")
		return true
	}

	// 3. SCORING SYSTEM
	score := 0

	if channel == ChannelSMS {
		score += dg.config.ScoreChannelSMS
	}

	if tone == ToneApologetic || tone == ToneProactive {
		score += dg.config.ScoreToneApologeticProactive
	}
	if len(input.LikelyReason) > dg.config.LikelyReasonLengthThreshold {
		score += dg.config.ScoreLongReason
	}

	needAI := score >= dg.config.AIScoreThreshold

	fmt.Printf("[DEBUG][shouldCallAI] ExceptionType: %s | Tone: %q | Channel: %q | Score: %d/%d | NeedAI: %v\n",
		input.ExceptionType, tone, channel, score, dg.config.AIScoreThreshold, needAI)

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
