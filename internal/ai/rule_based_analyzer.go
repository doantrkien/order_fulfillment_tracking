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

type RuleFunc func(aiCtx *models.AIContext, now time.Time) *RuleBasedResult

var rules = []RuleFunc{
	func(ctx *models.AIContext, now time.Time) *RuleBasedResult { return detectDuplicateEvents(ctx.Events) },
	func(ctx *models.AIContext, now time.Time) *RuleBasedResult {
		return detectInvalidTransitions(ctx.Events)
	},
	func(ctx *models.AIContext, now time.Time) *RuleBasedResult { return detectSkippedStatuses(ctx.Events) },
	func(ctx *models.AIContext, now time.Time) *RuleBasedResult { return detectStuckOrder(ctx, now) },
	func(ctx *models.AIContext, now time.Time) *RuleBasedResult { return detectDeliveryFailure(ctx) },
	func(ctx *models.AIContext, now time.Time) *RuleBasedResult { return detectCancellationAnomaly(ctx) },
	func(ctx *models.AIContext, now time.Time) *RuleBasedResult { return detectRefundAnomaly(ctx) },

	detectHealthyDelivered,
}

func AnalyzeByRules(aiCtx *models.AIContext, now time.Time) *RuleBasedResult {
	for _, rule := range rules {
		if r := rule(aiCtx, now); r != nil {
			return r
		}
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

func detectHealthyDelivered(aiCtx *models.AIContext, now time.Time) *RuleBasedResult {
	if aiCtx.CurrentStatus != models.ORDER_STATUS_DELIVERED {
		return nil
	}

	// If there are driver notes, this is an alternative delivery (deviation from
	// happy path). Return nil so AI can analyze and classify as ALTERNATIVE_DELIVERY.
	for _, e := range aiCtx.Events {
		if e.DriverNote != nil && strings.TrimSpace(*e.DriverNote) != "" {
			return nil
		}
	}

	// No driver notes — truly a happy-path delivery, no exception to report.
	return &RuleBasedResult{
		ExceptionType:      "NONE",
		Severity:           "LOW",
		LikelyReason:       "Order delivered successfully with no abnormal events.",
		InternalNextAction: "Close order, no action required.",
		ConfidenceScore:    1.0,
	}
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

		prevIdx, prevOk := statusIdx[e.PreviousStatus]
		newIdx, newOk := statusIdx[e.NewStatus]

		if !prevOk || !newOk {
			continue
		}

		skipCount := newIdx - prevIdx - 1

		if skipCount > 0 {
			switch {
			case skipCount == 1:
				return &RuleBasedResult{
					ExceptionType:      "SKIPPED_STATUS",
					Severity:           "MEDIUM",
					LikelyReason:       fmt.Sprintf("Order transitioned directly from '%s' to '%s', skipping one required status.", e.PreviousStatus, e.NewStatus),
					InternalNextAction: "Review order event generation.",
					ConfidenceScore:    1,
				}

			case skipCount == 2:
				return &RuleBasedResult{
					ExceptionType:      "SKIPPED_STATUS",
					Severity:           "HIGH",
					LikelyReason:       fmt.Sprintf("Order skipped multiple statuses: '%s' -> '%s'.", e.PreviousStatus, e.NewStatus),
					InternalNextAction: "Investigate workflow integrity.",
					ConfidenceScore:    1,
				}

			default:
				return &RuleBasedResult{
					ExceptionType:      "SKIPPED_STATUS",
					Severity:           "CRITICAL",
					LikelyReason:       fmt.Sprintf("Order skipped several mandatory statuses: '%s' -> '%s'.", e.PreviousStatus, e.NewStatus),
					InternalNextAction: "Immediate investigation required.",
					ConfidenceScore:    1,
				}
			}
		}
	}

	for i := 1; i < len(events); i++ {

		prevEvent := events[i-1]
		currEvent := events[i]

		if prevEvent.NewStatus == "" || currEvent.PreviousStatus == "" {
			continue
		}

		if prevEvent.NewStatus != currEvent.PreviousStatus {

			prevIdx, ok1 := statusIdx[prevEvent.NewStatus]
			currIdx, ok2 := statusIdx[currEvent.PreviousStatus]

			if !ok1 || !ok2 {
				continue
			}

			if currIdx > prevIdx {
				return &RuleBasedResult{
					ExceptionType: "SKIPPED_STATUS",
					Severity:      "MEDIUM",
					LikelyReason: fmt.Sprintf(
						"Missing event detected between '%s' and '%s'.",
						prevEvent.NewStatus,
						currEvent.PreviousStatus,
					),
					InternalNextAction: "Verify missing order events or event ingestion pipeline.",
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

var cancellationSeverity = map[models.OrderStatus]string{
	models.ORDER_STATUS_CREATED: "LOW",
}

var refundSeverity = map[models.OrderStatus]string{
	models.ORDER_STATUS_PAID:      "LOW",
	models.ORDER_STATUS_PACKED:    "MEDIUM",
	models.ORDER_STATUS_SHIPPED:   "HIGH",
	models.ORDER_STATUS_DELIVERED: "CRITICAL",
}

func detectCancellationAnomaly(aiCtx *models.AIContext) *RuleBasedResult {
	if aiCtx.CurrentStatus != models.ORDER_STATUS_CANCELLED {
		return nil
	}

	var prevStatus models.OrderStatus
	for _, e := range aiCtx.Events {
		if e.NewStatus == models.ORDER_STATUS_CANCELLED {
			prevStatus = e.PreviousStatus
			break
		}
	}

	severity, ok := cancellationSeverity[prevStatus]
	if !ok {
		severity = "LOW"
	}

	var reason, action string
	switch severity {
	case "LOW":
		reason = fmt.Sprintf("Order was cancelled early at stage '%s', minimal operational impact", prevStatus)
		action = "Log the cancellation and notify the customer support team for record keeping"
	case "MEDIUM":
		reason = fmt.Sprintf("Order was cancelled after payment at stage '%s', refund processing may be required", prevStatus)
		action = "Verify refund status and notify finance team to process any pending refund"
	case "HIGH":
		reason = fmt.Sprintf("Order was cancelled after packing at stage '%s', warehouse resources were already consumed", prevStatus)
		action = "Notify warehouse to restock items and finance to process refund; review cancellation policy"
	case "CRITICAL":
		reason = fmt.Sprintf("Order was cancelled after shipping at stage '%s', recall from carrier required", prevStatus)
		action = "Immediately contact the carrier to intercept and return the shipment; escalate to operations manager"
	}

	return &RuleBasedResult{
		ExceptionType:      "CANCELLATION_ANOMALY",
		Severity:           severity,
		LikelyReason:       reason,
		InternalNextAction: action,
		ConfidenceScore:    1.0,
	}
}

func detectRefundAnomaly(aiCtx *models.AIContext) *RuleBasedResult {
	if aiCtx.CurrentStatus != models.ORDER_STATUS_REFUNDED {
		return nil
	}

	var prevStatus models.OrderStatus
	for _, e := range aiCtx.Events {
		if e.NewStatus == models.ORDER_STATUS_REFUNDED {
			prevStatus = e.PreviousStatus
			break
		}
	}

	severity, ok := refundSeverity[prevStatus]
	if !ok {
		severity = "LOW"
	}

	var reason, action string
	switch severity {
	case "LOW":
		reason = fmt.Sprintf("Refund was issued early at stage '%s', likely a payment error or duplicate charge", prevStatus)
		action = "Verify payment records and confirm the refund amount with the finance team"
	case "MEDIUM":
		reason = fmt.Sprintf("Refund issued at stage '%s' while items were already packed, logistics coordination needed", prevStatus)
		action = "Notify warehouse to halt packing/dispatch and confirm refund with finance team"
	case "HIGH":
		reason = fmt.Sprintf("Refund issued at stage '%s' while order is in transit, carrier recall required", prevStatus)
		action = "Contact carrier to intercept shipment; coordinate with finance for refund and logistics for return"
	case "CRITICAL":
		reason = fmt.Sprintf("Refund issued at stage '%s' after successful delivery, possible fraud or system error", prevStatus)
		action = "Freeze the refund transaction, escalate to fraud/risk team, and initiate investigation immediately"
	}

	return &RuleBasedResult{
		ExceptionType:      "REFUND_ANOMALY",
		Severity:           severity,
		LikelyReason:       reason,
		InternalNextAction: action,
		ConfidenceScore:    1.0,
	}
}
