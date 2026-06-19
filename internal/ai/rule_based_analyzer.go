package ai

import (
	"fmt"
	"main/internal/models"
	"strings"
	"time"
)

// RuleBasedResult holds the output of the deterministic rule-based analyzer.
type RuleBasedResult struct {
	ExceptionType      string
	Severity           string
	LikelyReason       string
	InternalNextAction string
	ConfidenceScore    float64 // Always 1.0 for deterministic rules
}

// stuckThreshold defines per-status thresholds for stuck order detection.
type stuckThreshold struct {
	Duration         time.Duration
	BaseSeverity     string
	EscalateSeverity string // severity when age > 2× threshold
}

// stuckThresholds maps each non-terminal status to its stuck-order threshold.
var stuckThresholds = map[models.OrderStatus]stuckThreshold{
	models.ORDER_STATUS_CREATED: {
		Duration:         24 * time.Hour,
		BaseSeverity:     "MEDIUM",
		EscalateSeverity: "CRITICAL",
	},
	models.ORDER_STATUS_PAID: {
		Duration:         48 * time.Hour,
		BaseSeverity:     "MEDIUM",
		EscalateSeverity: "CRITICAL",
	},
	models.ORDER_STATUS_PACKED: {
		Duration:         24 * time.Hour,
		BaseSeverity:     "HIGH",
		EscalateSeverity: "CRITICAL",
	},
	models.ORDER_STATUS_SHIPPED: {
		Duration:         72 * time.Hour,
		BaseSeverity:     "HIGH",
		EscalateSeverity: "CRITICAL",
	},
}

// lifecycleOrder defines the expected forward path for the order lifecycle.
var lifecycleOrder = []models.OrderStatus{
	models.ORDER_STATUS_CREATED,
	models.ORDER_STATUS_PAID,
	models.ORDER_STATUS_PACKED,
	models.ORDER_STATUS_SHIPPED,
	models.ORDER_STATUS_DELIVERED,
}

// deliveryFailureKeywords are keywords indicating a severe delivery failure (HIGH severity).
// "not home", "customer not home" → MEDIUM severity (see mediumDeliveryKeywords)
var deliveryFailureKeywords = []string{
	"weather", "bad weather", "weather condition",
	"accident", "vehicle breakdown", "xe hỏng", "tai nạn",
	"failed", "failure", "lost", "package lost",
	"cannot deliver", "could not deliver",
	"giao thất bại", "thời tiết",
}

// mediumDeliveryKeywords are keywords indicating a softer delivery failure (MEDIUM severity).
var mediumDeliveryKeywords = []string{
	"not home", "customer not home", "no one home",
	"wrong address", "address not found",
	"không có nhà", "sai địa chỉ",
}

// cancellationAfterShippedStatuses lists statuses from which cancellation is anomalous.
var cancellationAfterShippedStatuses = map[models.OrderStatus]bool{
	models.ORDER_STATUS_SHIPPED:   true,
	models.ORDER_STATUS_DELIVERED: true,
}

// AnalyzeByRules runs deterministic rule-based exception detection on the order context.
// Rules are evaluated in priority order. The first matching rule wins.
// Returns nil if no exception is detected.
func AnalyzeByRules(aiCtx *models.AIContext, now time.Time) *RuleBasedResult {
	// Priority 1: Cancellation anomaly (CRITICAL)
	if r := detectCancellationAnomaly(aiCtx); r != nil {
		return r
	}

	// Priority 2: Refund anomaly (HIGH/CRITICAL)
	if r := detectRefundAnomaly(aiCtx); r != nil {
		return r
	}

	// Priority 3: Delivery failure (MEDIUM/HIGH)
	if r := detectDeliveryFailure(aiCtx); r != nil {
		return r
	}

	// // Priority 4: Duplicate events (LOW)
	// if r := detectDuplicateEvents(aiCtx.Events); r != nil {
	// 	return r
	// }

	// Priority 5: Skipped statuses (HIGH)
	if r := detectSkippedStatuses(aiCtx.Events); r != nil {
		return r
	}

	// Priority 6: Invalid transitions (CRITICAL)
	if r := detectInvalidTransitions(aiCtx.Events); r != nil {
		return r
	}

	// Priority 7: Stuck order (severity varies by age)
	if r := detectStuckOrder(aiCtx, now); r != nil {
		return r
	}

	return nil
}

// detectInvalidTransitions checks if any event in the timeline has
// a from→to transition that is not in the valid transitions map.
func detectInvalidTransitions(events []models.AIEvent) *RuleBasedResult {
	for _, e := range events {
		if e.PreviousStatus == e.NewStatus {
			continue // Handled by detectDuplicateEvents
		}
		if !models.IsValidTransition(e.PreviousStatus, e.NewStatus) {
			return &RuleBasedResult{
				ExceptionType:      "INVALID_TRANSITION",
				Severity:           "CRITICAL",
				LikelyReason:       fmt.Sprintf("Invalid status transition from '%s' to '%s'", e.PreviousStatus, e.NewStatus),
				InternalNextAction: "Review the event source and block further invalid transitions. Escalate to engineering if automated.",
				ConfidenceScore:    1.0,
			}
		}
	}
	return nil
}

// detectDuplicateEvents checks for consecutive events with the same new_status.
func detectDuplicateEvents(events []models.AIEvent) *RuleBasedResult {
	for i := 1; i < len(events); i++ {
		if events[i].NewStatus == events[i-1].NewStatus {
			return &RuleBasedResult{
				ExceptionType:      "DUPLICATE_EVENT",
				Severity:           "LOW",
				LikelyReason:       fmt.Sprintf("Duplicate consecutive event detected: status '%s' recorded multiple times", events[i].NewStatus),
				InternalNextAction: "Investigate the event source for duplicate submissions. No immediate action required.",
				ConfidenceScore:    1.0,
			}
		}
	}
	return nil
}

// detectSkippedStatuses walks the event timeline and checks for skipped
// statuses in the expected lifecycle path (created → paid → packed → shipped → delivered).
func detectSkippedStatuses(events []models.AIEvent) *RuleBasedResult {
	if len(events) == 0 {
		return nil
	}

	seen := make(map[models.OrderStatus]bool)
	for _, e := range events {
		if e.PreviousStatus != "" {
			seen[e.PreviousStatus] = true
		}
		if e.NewStatus != "" {
			seen[e.NewStatus] = true
		}
	}

	// Find the furthest lifecycle stage reached
	maxIdx := -1
	for i, status := range lifecycleOrder {
		if seen[status] {
			maxIdx = i
		}
	}

	if maxIdx <= 0 {
		// No meaningful forward progress to check skips
		return nil
	}

	// Check if any intermediate status was skipped
	for i := 1; i < maxIdx; i++ {
		if !seen[lifecycleOrder[i]] {
			return &RuleBasedResult{
				ExceptionType:      "SKIPPED_STATUS",
				Severity:           "HIGH",
				LikelyReason:       fmt.Sprintf("Status '%s' was skipped in the order lifecycle", lifecycleOrder[i]),
				InternalNextAction: "Verify the order processing pipeline. Missing status may indicate a system bypass or data integrity issue.",
				ConfidenceScore:    1.0,
			}
		}
	}

	return nil
}

// detectStuckOrder calculates the time since the last event (or order creation)
// and compares against per-status thresholds.
func detectStuckOrder(aiCtx *models.AIContext, now time.Time) *RuleBasedResult {
	currentStatus := aiCtx.CurrentStatus

	threshold, exists := stuckThresholds[currentStatus]
	if !exists {
		// Terminal states (delivered, cancelled, refunded) cannot be "stuck"
		return nil
	}

	// Determine the last activity time
	lastActivityAt := aiCtx.CreatedAt
	if len(aiCtx.Events) > 0 {
		lastEvent := aiCtx.Events[len(aiCtx.Events)-1]
		if lastEvent.EventAt.After(lastActivityAt) {
			lastActivityAt = lastEvent.EventAt
		}
	}

	age := now.Sub(lastActivityAt)
	if age < threshold.Duration {
		return nil
	}

	severity := threshold.BaseSeverity
	if age >= 2*threshold.Duration {
		severity = threshold.EscalateSeverity
	}

	hours := int(age.Hours())

	return &RuleBasedResult{
		ExceptionType:      "STUCK_ORDER",
		Severity:           severity,
		LikelyReason:       fmt.Sprintf("Order has been in '%s' status for %d hours without progressing", currentStatus, hours),
		InternalNextAction: fmt.Sprintf("Investigate why order has not advanced from '%s'. Contact the responsible team.", currentStatus),
		ConfidenceScore:    1.0,
	}
}

// detectDeliveryFailure inspects DriverNotes for keywords that indicate a delivery failure.
// Severity is HIGH for severe causes (weather, accident, vehicle breakdown, lost)
// and MEDIUM for softer causes (customer not home, wrong address).
func detectDeliveryFailure(aiCtx *models.AIContext) *RuleBasedResult {
	// Only relevant when order is in "shipped" status
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

	// Check medium-severity keywords first (more specific)
	for _, kw := range mediumDeliveryKeywords {
		if strings.Contains(notes, kw) {
			return &RuleBasedResult{
				ExceptionType:      "DELIVERY_FAILURE",
				Severity:           "MEDIUM",
				LikelyReason:       fmt.Sprintf("Delivery failed: %s", rawNotes),
				InternalNextAction: "Contact customer to reschedule delivery. Update delivery attempts log.",
				ConfidenceScore:    1.0,
			}
		}
	}

	// Check high-severity keywords
	for _, kw := range deliveryFailureKeywords {
		if strings.Contains(notes, kw) {
			return &RuleBasedResult{
				ExceptionType:      "DELIVERY_FAILURE",
				Severity:           "HIGH",
				LikelyReason:       fmt.Sprintf("Delivery failed: %s", rawNotes),
				InternalNextAction: "Escalate to logistics team. Arrange re-delivery or return to warehouse.",
				ConfidenceScore:    1.0,
			}
		}
	}

	return nil
}

// detectCancellationAnomaly detects orders cancelled after being shipped or delivered.
// These are CRITICAL because they represent high-impact operational anomalies.
func detectCancellationAnomaly(aiCtx *models.AIContext) *RuleBasedResult {
	if aiCtx.CurrentStatus != models.ORDER_STATUS_CANCELLED {
		return nil
	}

	// Walk the event history looking for a transition INTO cancelled
	// from a late-stage status (shipped or delivered)
	for _, e := range aiCtx.Events {
		if e.NewStatus == models.ORDER_STATUS_CANCELLED {
			if cancellationAfterShippedStatuses[e.PreviousStatus] {
				return &RuleBasedResult{
					ExceptionType:      "CANCELLATION_ANOMALY",
					Severity:           "CRITICAL",
					LikelyReason:       fmt.Sprintf("Order was cancelled after reaching '%s' status — late-stage cancellation detected", e.PreviousStatus),
					InternalNextAction: "Halt any ongoing delivery. Initiate return-to-warehouse procedure. Review refund eligibility.",
					ConfidenceScore:    1.0,
				}
			}
		}
	}

	return nil
}

// detectRefundAnomaly detects anomalous refund scenarios:
// - CRITICAL: refunded directly from "created" (order was never paid)
// - HIGH: refunded directly from "paid" (skipped normal cancellation flow, potential double refund)
func detectRefundAnomaly(aiCtx *models.AIContext) *RuleBasedResult {
	if aiCtx.CurrentStatus != models.ORDER_STATUS_REFUNDED {
		return nil
	}

	// Walk the event history looking for the transition INTO refunded
	for _, e := range aiCtx.Events {
		if e.NewStatus == models.ORDER_STATUS_REFUNDED {
			switch e.PreviousStatus {
			case models.ORDER_STATUS_CREATED:
				// Refunded from created = order was never paid
				return &RuleBasedResult{
					ExceptionType:      "REFUND_ANOMALY",
					Severity:           "CRITICAL",
					LikelyReason:       "Refund was processed for an order that was never paid",
					InternalNextAction: "Immediately investigate payment gateway logs. Reverse the refund transaction if fraudulent. Escalate to finance team.",
					ConfidenceScore:    1.0,
				}
			case models.ORDER_STATUS_PAID:
				// Refunded directly from paid without going through cancellation
				return &RuleBasedResult{
					ExceptionType:      "REFUND_ANOMALY",
					Severity:           "HIGH",
					LikelyReason:       "Refund was processed directly from paid status — potential double refund or bypassed cancellation flow",
					InternalNextAction: "Verify refund legitimacy with payment team. Check for duplicate refund requests. Audit payment gateway records.",
					ConfidenceScore:    1.0,
				}
			}
		}
	}

	return nil
}
