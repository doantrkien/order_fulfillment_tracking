package ai

import (
	"fmt"
	"main/internal/models"
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

// AnalyzeByRules runs deterministic rule-based exception detection on the order context.
// Rules are evaluated in priority order. The first matching rule wins.
// Returns nil if no exception is detected.
func AnalyzeByRules(aiCtx *models.AIContext, now time.Time) *RuleBasedResult {
	// Priority 1: Invalid transitions (CRITICAL)
	if r := detectInvalidTransitions(aiCtx.Events); r != nil {
		return r
	}

	// Priority 2: Duplicate events (LOW)
	if r := detectDuplicateEvents(aiCtx.Events); r != nil {
		return r
	}

	// Priority 3: Skipped statuses (HIGH)
	if r := detectSkippedStatuses(aiCtx.Events); r != nil {
		return r
	}

	// Priority 4: Stuck order (severity varies by age)
	if r := detectStuckOrder(aiCtx, now); r != nil {
		return r
	}

	return nil
}

// detectInvalidTransitions checks if any event in the timeline has
// a from→to transition that is not in the valid transitions map.
func detectInvalidTransitions(events []models.AIEvent) *RuleBasedResult {
	fmt.Printf("[DEBUG][detectInvalidTransitions] Checking events for invalid transitions\n", events)
	for _, e := range events {
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

	// Collect all statuses that have appeared as new_status in the timeline
	seen := make(map[models.OrderStatus]bool)
	for _, e := range events {
		seen[e.NewStatus] = true
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
