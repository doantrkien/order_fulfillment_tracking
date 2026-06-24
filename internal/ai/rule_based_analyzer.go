package ai

import (
	"fmt"
	"main/internal/models"
	"strings"
	"time"
)

type RuleBasedResult struct {
	ExceptionType      string
	Severity           string
	LikelyReason       string
	InternalNextAction string
	ConfidenceScore    float64
}

var stuckThresholds = map[models.OrderStatus]time.Duration{
	models.ORDER_STATUS_CREATED: 24 * time.Hour,
	models.ORDER_STATUS_PAID:    48 * time.Hour,
	models.ORDER_STATUS_PACKED:  24 * time.Hour,
	models.ORDER_STATUS_SHIPPED: 72 * time.Hour,
}

var lifecycleOrder = []models.OrderStatus{
	models.ORDER_STATUS_CREATED,
	models.ORDER_STATUS_PAID,
	models.ORDER_STATUS_PACKED,
	models.ORDER_STATUS_SHIPPED,
	models.ORDER_STATUS_DELIVERED,
}

var deliveryFailureCriticalKeywords = []string{
	"package lost", "stolen", "cannot find package", "missing parcel",
	"hàng bị mất", "không tìm thấy kiện hàng", "nghi thất lạc",
}

var deliveryFailureHighKeywords = []string{
	"vehicle breakdown", "accident on route", "bad weather", "road blocked",
	"xe hỏng", "tai nạn giao thông", "thời tiết xấu",
}

var deliveryFailureMediumKeywords = []string{
	"customer not home", "no one available to receive package", "wrong address", "address not found",
	"không có ai ở nhà", "sai địa chỉ", "không liên lạc được khách hàng",
}

var deliveryFailureLowKeywords = []string{
	"customer temporarily unreachable", "no answer", "will retry call", "short delay at delivery point",
	"khách không nghe máy tạm thời",
}

func AnalyzeByRules(aiCtx *models.AIContext, now time.Time) *RuleBasedResult {
	if r := detectDeliveryFailure(aiCtx); r != nil {
		return r
	}
	if r := detectDuplicateEvents(aiCtx.Events); r != nil {
		return r
	}
	if r := detectSkippedStatuses(aiCtx.Events); r != nil {
		return r
	}
	if r := detectInvalidTransitions(aiCtx.Events); r != nil {
		return r
	}
	if r := detectStuckOrder(aiCtx, now); r != nil {
		return r
	}
	return nil
}

func detectDeliveryFailure(aiCtx *models.AIContext) *RuleBasedResult {
	if aiCtx.CurrentStatus != models.ORDER_STATUS_SHIPPED {
		return nil
	}

	var notesParts []string
	for _, e := range aiCtx.Events {
		if e.DriverNote != nil && strings.TrimSpace(*e.DriverNote) != "" {
			notesParts = append(notesParts, strings.TrimSpace(*e.DriverNote))
		}
	}
	rawNotes := strings.Join(notesParts, "; ")
	notes := strings.ToLower(rawNotes)
	if notes == "" {
		return nil
	}

	for _, kw := range deliveryFailureCriticalKeywords {
		if strings.Contains(notes, kw) {
			return &RuleBasedResult{
				ExceptionType:      "DELIVERY_FAILURE",
				Severity:           "CRITICAL",
				LikelyReason:       "Package is missing or potentially lost in transit",
				InternalNextAction: "Escalate to logistics manager. Open lost-parcel investigation and notify finance for claim handling.",
				ConfidenceScore:    1,
			}
		}
	}

	for _, kw := range deliveryFailureHighKeywords {
		if strings.Contains(notes, kw) {
			return &RuleBasedResult{
				ExceptionType:      "DELIVERY_FAILURE",
				Severity:           "HIGH",
				LikelyReason:       "Operational issue prevented delivery",
				InternalNextAction: "Escalate to logistics team. Arrange re-delivery or return-to-warehouse.",
				ConfidenceScore:    1,
			}
		}
	}

	for _, kw := range deliveryFailureMediumKeywords {
		if strings.Contains(notes, kw) {
			return &RuleBasedResult{
				ExceptionType:      "DELIVERY_FAILURE",
				Severity:           "MEDIUM",
				LikelyReason:       "Customer not available at delivery location or address issue",
				InternalNextAction: "Contact customer to reschedule delivery and log attempt",
				ConfidenceScore:    1,
			}
		}
	}

	for _, kw := range deliveryFailureLowKeywords {
		if strings.Contains(notes, kw) {
			return &RuleBasedResult{
				ExceptionType:      "DELIVERY_FAILURE",
				Severity:           "LOW",
				LikelyReason:       "Customer temporarily unreachable but delivery can be retried immediately",
				InternalNextAction: "Retry contact customer and reattempt delivery",
				ConfidenceScore:    1,
			}
		}
	}

	return nil
}

func detectDuplicateEvents(events []models.AIEvent) *RuleBasedResult {
	seen := make(map[models.OrderStatus]bool)
	for _, e := range events {
		if e.NewStatus != "" {
			if seen[e.NewStatus] {
				switch e.NewStatus {
				case models.ORDER_STATUS_DELIVERED:
					return &RuleBasedResult{
						ExceptionType:      "DUPLICATE_EVENT",
						Severity:           "CRITICAL",
						LikelyReason:       "Duplicate event caused incorrect financial operation: refund executed twice",
						InternalNextAction: "Stop processing, rollback incorrect transactions, and audit event pipeline",
						ConfidenceScore:    1,
					}
				case models.ORDER_STATUS_SHIPPED:
					return &RuleBasedResult{
						ExceptionType:      "DUPLICATE_EVENT",
						Severity:           "HIGH",
						LikelyReason:       "Duplicate event caused repeated external side effects (shipping notification)",
						InternalNextAction: "Fix idempotency in downstream services and add deduplication layer",
						ConfidenceScore:    1,
					}
				case models.ORDER_STATUS_PAID:
					return &RuleBasedResult{
						ExceptionType:      "DUPLICATE_EVENT",
						Severity:           "MEDIUM",
						LikelyReason:       "Duplicate event triggered unnecessary internal reprocessing for status 'paid'",
						InternalNextAction: "Investigate consumer idempotency and reduce redundant processing",
						ConfidenceScore:    1,
					}
				default:
					return &RuleBasedResult{
						ExceptionType:      "DUPLICATE_EVENT",
						Severity:           "LOW",
						LikelyReason:       fmt.Sprintf("Duplicate event detected: status '%s' received more than once with no side effect", e.NewStatus),
						InternalNextAction: "Log and monitor event source for retry behavior",
						ConfidenceScore:    1,
					}
				}
			}
			seen[e.NewStatus] = true
		}
	}
	return nil
}

func detectSkippedStatuses(events []models.AIEvent) *RuleBasedResult {
	statusIdx := map[models.OrderStatus]int{
		models.ORDER_STATUS_CREATED:   0,
		models.ORDER_STATUS_PAID:      1,
		models.ORDER_STATUS_PACKED:    2,
		models.ORDER_STATUS_SHIPPED:   3,
		models.ORDER_STATUS_DELIVERED: 4,
	}

	for _, e := range events {
		if e.PreviousStatus == "" || e.NewStatus == "" {
			continue
		}

		if e.PreviousStatus == models.ORDER_STATUS_CREATED && e.NewStatus == models.ORDER_STATUS_REFUNDED {
			return &RuleBasedResult{
				ExceptionType:      "SKIPPED_STATUS",
				Severity:           "CRITICAL",
				LikelyReason:       "Order transitioned from 'created' directly to 'refunded', skipping required status 'paid'",
				InternalNextAction: "Perform immediate data integrity audit and investigate the processing pipeline",
				ConfidenceScore:    1,
			}
		}

		prevIdx, prevOk := statusIdx[e.PreviousStatus]
		newIdx, newOk := statusIdx[e.NewStatus]

		if prevOk && newOk {
			skipCount := newIdx - prevIdx - 1
			if skipCount == 1 {
				if e.PreviousStatus == models.ORDER_STATUS_CREATED && e.NewStatus == models.ORDER_STATUS_PACKED {
					return &RuleBasedResult{
						ExceptionType:      "SKIPPED_STATUS",
						Severity:           "LOW",
						LikelyReason:       "Order transitioned from 'created' directly to 'packed', skipping required status 'paid'",
						InternalNextAction: "Review event logs and monitor for additional anomalies",
						ConfidenceScore:    1,
					}
				} else {
					return &RuleBasedResult{
						ExceptionType:      "SKIPPED_STATUS",
						Severity:           "MEDIUM",
						LikelyReason:       fmt.Sprintf("Order transitioned from '%s' directly to '%s', skipping a required fulfillment status", e.PreviousStatus, e.NewStatus),
						InternalNextAction: "Investigate the order processing pipeline and verify status generation",
						ConfidenceScore:    1,
					}
				}
			} else if skipCount == 2 {
				return &RuleBasedResult{
					ExceptionType:      "SKIPPED_STATUS",
					Severity:           "HIGH",
					LikelyReason:       fmt.Sprintf("Order skipped multiple required statuses: '%s' to '%s'", e.PreviousStatus, e.NewStatus),
					InternalNextAction: "Escalate to the responsible engineering team and investigate workflow integrity",
					ConfidenceScore:    1,
				}
			} else if skipCount >= 3 {
				return &RuleBasedResult{
					ExceptionType:      "SKIPPED_STATUS",
					Severity:           "CRITICAL",
					LikelyReason:       fmt.Sprintf("Order transitioned from '%s' directly to '%s', skipping several mandatory lifecycle stages", e.PreviousStatus, e.NewStatus),
					InternalNextAction: "Perform immediate data integrity audit and investigate the processing pipeline",
					ConfidenceScore:    1,
				}
			}
		}
	}
	return nil
}

func detectInvalidTransitions(events []models.AIEvent) *RuleBasedResult {
	for _, e := range events {
		if e.PreviousStatus == "" || e.NewStatus == "" || e.PreviousStatus == e.NewStatus {
			continue
		}
		if !models.IsValidTransition(e.PreviousStatus, e.NewStatus) {
			return &RuleBasedResult{
				ExceptionType:      "INVALID_TRANSITION",
				Severity:           "CRITICAL",
				LikelyReason:       fmt.Sprintf("Attempted invalid transition from '%s' to '%s'", e.PreviousStatus, e.NewStatus),
				InternalNextAction: "Block processing immediately and perform data consistency checks",
				ConfidenceScore:    1,
			}
		}
	}
	return nil
}

func detectStuckOrder(aiCtx *models.AIContext, now time.Time) *RuleBasedResult {
	currentStatus := aiCtx.CurrentStatus

	threshold, exists := stuckThresholds[currentStatus]
	if !exists {
		return nil
	}

	lastActivityAt := aiCtx.CreatedAt
	if len(aiCtx.Events) > 0 {
		lastEvent := aiCtx.Events[len(aiCtx.Events)-1]
		if lastEvent.EventAt.After(lastActivityAt) {
			lastActivityAt = lastEvent.EventAt
		}
	}

	age := now.Sub(lastActivityAt)
	if age <= threshold {
		return nil
	}

	ratio := float64(age) / float64(threshold)
	hours := int(age.Hours())

	if ratio <= 1.25 {
		return &RuleBasedResult{
			ExceptionType:      "STUCK_ORDER",
			Severity:           "LOW",
			LikelyReason:       fmt.Sprintf("Order has been in '%s' status for %d hours, slightly exceeding the normal threshold", currentStatus, hours),
			InternalNextAction: "Monitor and notify the responsible team",
			ConfidenceScore:    1,
		}
	} else if ratio <= 2.0 {
		return &RuleBasedResult{
			ExceptionType:      "STUCK_ORDER",
			Severity:           "MEDIUM",
			LikelyReason:       fmt.Sprintf("Order has been in '%s' status for %d hours without progressing", currentStatus, hours),
			InternalNextAction: "Investigate the delay and contact the responsible team",
			ConfidenceScore:    1,
		}
	} else if ratio <= 3.0 {
		return &RuleBasedResult{
			ExceptionType:      "STUCK_ORDER",
			Severity:           "HIGH",
			LikelyReason:       fmt.Sprintf("Order has been in '%s' status for %d hours, exceeding 2× the normal threshold", currentStatus, hours),
			InternalNextAction: "Escalate to the team lead and investigate immediately",
			ConfidenceScore:    1,
		}
	} else {
		return &RuleBasedResult{
			ExceptionType:      "STUCK_ORDER",
			Severity:           "CRITICAL",
			LikelyReason:       fmt.Sprintf("Order has been in '%s' status for %d hours, exceeding 3× the normal threshold", currentStatus, hours),
			InternalNextAction: "Urgently escalate to operations management and investigate immediately",
			ConfidenceScore:    1,
		}
	}
}
