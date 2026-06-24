package ai

import (
	"regexp"
	"strings"
)

type DriverNoteCategory string

const (
	DriverNoteCategoryDeliveryFailure DriverNoteCategory = "DELIVERY_FAILURE"
	DriverNoteCategoryDelay           DriverNoteCategory = "DELAY"
	DriverNoteCategoryCancellation    DriverNoteCategory = "CANCELLATION"
	DriverNoteCategoryGeneral         DriverNoteCategory = "GENERAL"
	DriverNoteCategoryNone            DriverNoteCategory = "NONE"
)

type DriverNoteClassification struct {
	Category  DriverNoteCategory
	KBSnippet string
}

// ClassifyDriverNote analyzes driver notes using deterministic keyword/regex rules
// and returns the appropriate Knowledge Base snippet to inject into the prompt,
// along with the category for Orchestrator routing.
// This is intentionally NON-AI to avoid double LLM calls (latency + cost).
func ClassifyDriverNote(note string) DriverNoteClassification {
	if strings.TrimSpace(note) == "" {
		return DriverNoteClassification{Category: DriverNoteCategoryNone, KBSnippet: ""}
	}

	lowerNote := strings.ToLower(note)

	// Rule 1: Delivery Failure
	// Keywords: "hỏng", "tai nạn", "không liên lạc được", "vắng mặt", "từ chối"
	deliveryFailureRegex := regexp.MustCompile(`(hỏng|tai nạn|không liên lạc được|vắng mặt|từ chối|sai địa chỉ)`)
	if deliveryFailureRegex.MatchString(lowerNote) {
		return DriverNoteClassification{
			Category: DriverNoteCategoryDeliveryFailure,
			KBSnippet: `[SPECIFIC KNOWLEDGE: DELIVERY FAILURE]
- If the driver notes indicate an inability to contact the customer, address issues, or refusal, classify as DELIVERY_FAILURE.
- Severity is typically HIGH.
- Next action should focus on contacting the customer to verify details or arranging a redelivery.`,
		}
	}

	// Rule 2: Stuck/Delay
	// Keywords: "chờ", "delay", "tắc đường", "trễ", "thời tiết"
	delayRegex := regexp.MustCompile(`(chờ|delay|tắc đường|trễ|thời tiết|mưa|bão)`)
	if delayRegex.MatchString(lowerNote) {
		return DriverNoteClassification{
			Category: DriverNoteCategoryDelay,
			KBSnippet: `[SPECIFIC KNOWLEDGE: DELAYED TRANSIT]
- If the driver notes indicate traffic, weather delays, or waiting times, it often leads to a STUCK_ORDER or delayed fulfillment.
- Severity depends on the age of the delay (MEDIUM to HIGH).
- Next action should involve monitoring and notifying operations if it exceeds thresholds.`,
		}
	}

	// Rule 3: Cancellation/Refund Anomaly
	// Keywords: "hủy", "trả hàng", "hoàn tiền"
	cancellationRegex := regexp.MustCompile(`(hủy|trả hàng|hoàn tiền|refund)`)
	if cancellationRegex.MatchString(lowerNote) {
		return DriverNoteClassification{
			Category: DriverNoteCategoryCancellation,
			KBSnippet: `[SPECIFIC KNOWLEDGE: CANCELLATION / REFUND]
- Mentions of returning goods, cancellations, or refunds in transit can indicate a CANCELLATION_ANOMALY or REFUND_ANOMALY.
- Check against current state. If order is shipped, severity is HIGH.
- Next action should route to customer support for return logistics.`,
		}
	}

	// Default: Generic guidance if no specific keywords match
	return DriverNoteClassification{
		Category: DriverNoteCategoryGeneral,
		KBSnippet: `[SPECIFIC KNOWLEDGE: GENERAL ANOMALY]
- The operator provided notes that do not strictly match standard delay or failure patterns.
- Analyze the timeline carefully to deduce the appropriate exception type (often OTHER or STUCK_ORDER).`,
	}
}
