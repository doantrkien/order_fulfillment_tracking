package ai

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed knowledge/state_machine.md
var kbBodyStateMachine string

//go:embed knowledge/delivery_failure.md
var kbBodyDeliveryFailure string

//go:embed knowledge/stuck_order.md
var kbBodyStuckOrder string

//go:embed knowledge/duplicate_event.md
var kbBodyDuplicateEvent string

//go:embed knowledge/skipped_status.md
var kbBodySkippedStatus string

type KnowledgeEntry struct {
	Title string
	Body  string
}

// ── Compiled KB entries (one per exception domain) ───────────────────────────
var kbStateMachine = KnowledgeEntry{
	Title: "Order State Machine",
	Body:  kbBodyStateMachine,
}

var kbDeliveryFailure = KnowledgeEntry{
	Title: "Delivery Failure",
	Body:  kbBodyDeliveryFailure,
}

var kbStuckOrder = KnowledgeEntry{
	Title: "Stuck Order",
	Body:  kbBodyStuckOrder,
}

var kbDuplicateEvent = KnowledgeEntry{
	Title: "Duplicate Event",
	Body:  kbBodyDuplicateEvent,
}

var kbSkippedStatus = KnowledgeEntry{
	Title: "Skipped Status",
	Body:  kbBodySkippedStatus,
}

var deliveryFailureKeywordsKB = []string{
	"lost", "stolen", "cannot find", "missing parcel", "hàng bị mất", "nghi thất lạc",

	"accident", "vehicle breakdown", "xe hỏng", "tai nạn", "bad weather", "thời tiết",
	"failed", "failure", "cannot deliver", "could not deliver", "giao thất bại",

	"not home", "no one home", "customer not home", "wrong address", "address not found",
	"không có nhà", "sai địa chỉ", "không liên lạc được",

	"temporarily unreachable", "no answer", "will retry", "khách không nghe máy",
}

var duplicateKeywordsKB = []string{
	"duplicate", "lặp", "trùng", "repeated",
}

var skippedKeywordsKB = []string{
	"skipped", "missed", "bỏ qua", "thiếu",
}

var stuckKeywordsKB = []string{
	"stuck", "no progress", "not moved", "delay", "delayed", "late",
	"trễ", "chậm", "không tiến triển",
}

func ClassifyDriverNote(note string) []KnowledgeEntry {
	lower := strings.ToLower(strings.TrimSpace(note))
	entries := []KnowledgeEntry{kbStateMachine} // always present

	matched := false
	if containsAny(lower, deliveryFailureKeywordsKB) {
		entries = append(entries, kbDeliveryFailure)
		matched = true
	}
	if containsAny(lower, duplicateKeywordsKB) {
		entries = append(entries, kbDuplicateEvent)
		matched = true
	}
	if containsAny(lower, skippedKeywordsKB) {
		entries = append(entries, kbSkippedStatus)
		matched = true
	}
	if containsAny(lower, stuckKeywordsKB) {
		entries = append(entries, kbStuckOrder)
		matched = true
	}

	if !matched {
		entries = append(entries, kbDeliveryFailure, kbDuplicateEvent, kbSkippedStatus, kbStuckOrder)
	}

	var titles []string
	for _, e := range entries {
		titles = append(titles, e.Title)
	}
	fmt.Printf("[DEBUG][ClassifyDriverNote] Note: %q | Matched: %v | KB Sent: %d (%s)\n", note, matched, len(entries), strings.Join(titles, ", "))

	return entries
}

func containsAny(s string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
