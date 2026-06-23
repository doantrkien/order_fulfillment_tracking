package ai

import (
	_ "embed"
	"strings"
)

// Knowledge base files are stored as Markdown under the knowledge/ directory.
// Go's embed directive loads them at compile time — no runtime file I/O needed.
// To update a rule, edit the corresponding .md file; no Go code change required.

//go:embed knowledge/state_machine.md
var kbBodyStateMachine string

//go:embed knowledge/delivery_failure.md
var kbBodyDeliveryFailure string

//go:embed knowledge/cancellation_anomaly.md
var kbBodyCancellationAnomaly string

//go:embed knowledge/refund_anomaly.md
var kbBodyRefundAnomaly string

//go:embed knowledge/stuck_order.md
var kbBodyStuckOrder string

//go:embed knowledge/data_integrity.md
var kbBodyDataIntegrity string

// KnowledgeEntry represents a single knowledge base chunk injected into the prompt.
// Each entry covers one exception domain and contains the rules the AI needs
// to produce consistent, deterministic severity decisions for that domain.
type KnowledgeEntry struct {
	// ID is a short machine-readable identifier (e.g. "delivery_failure").
	ID string
	// Title is the section header shown inside the prompt.
	Title string
	// Body is the full knowledge text (sourced from the embedded .md file).
	Body string
}

// ── Compiled KB entries (one per exception domain) ───────────────────────────

// kbStateMachine covers the valid order lifecycle and all allowed transitions.
// Always injected — the AI must know the state machine regardless of context.
var kbStateMachine = KnowledgeEntry{
	ID:    "state_machine",
	Title: "Order State Machine",
	Body:  kbBodyStateMachine,
}

var kbDeliveryFailure = KnowledgeEntry{
	ID:    "delivery_failure",
	Title: "Delivery Failure",
	Body:  kbBodyDeliveryFailure,
}

var kbCancellationAnomaly = KnowledgeEntry{
	ID:    "cancellation_anomaly",
	Title: "Cancellation Anomaly",
	Body:  kbBodyCancellationAnomaly,
}

var kbRefundAnomaly = KnowledgeEntry{
	ID:    "refund_anomaly",
	Title: "Refund Anomaly",
	Body:  kbBodyRefundAnomaly,
}

var kbStuckOrder = KnowledgeEntry{
	ID:    "stuck_order",
	Title: "Stuck Order",
	Body:  kbBodyStuckOrder,
}

var kbDataIntegrity = KnowledgeEntry{
	ID:    "data_integrity",
	Title: "Data Integrity Issues",
	Body:  kbBodyDataIntegrity,
}

// ── Keyword sets used by ClassifyDriverNote ──────────────────────────────────

var deliveryFailureKeywordsKB = []string{
	// critical
	"lost", "stolen", "cannot be found",
	// high
	"accident", "vehicle breakdown", "xe hỏng", "tai nạn",
	"failed", "failure", "cannot deliver", "could not deliver",
	"giao thất bại",
	"weather", "bad weather", "thời tiết",
	// medium
	"not home", "no one home", "customer not home",
	"wrong address", "address not found",
	"không có nhà", "sai địa chỉ",
}

var cancellationKeywordsKB = []string{
	"cancel", "cancelled", "cancellation", "hủy", "huỷ",
}

var refundKeywordsKB = []string{
	"refund", "hoàn tiền", "hoàn trả", "trả hàng",
}

var stuckKeywordsKB = []string{
	"stuck", "no progress", "not moved", "delay", "delayed", "late",
	"trễ", "chậm", "không tiến triển",
}

// ── Public API ───────────────────────────────────────────────────────────────

// ClassifyDriverNote inspects the free-text driver note and returns the set of
// KnowledgeEntry items that are relevant to inject into the prompt.
//
// kbStateMachine is ALWAYS included as a baseline.
// Additional entries are selected by keyword matching against the note text.
// If no specific keywords match, delivery_failure + data_integrity are added as
// a sensible fallback so the AI has useful context for any driver-note scenario.
func ClassifyDriverNote(note string) []KnowledgeEntry {
	lower := strings.ToLower(strings.TrimSpace(note))
	entries := []KnowledgeEntry{kbStateMachine} // always present

	if containsAny(lower, deliveryFailureKeywordsKB) {
		entries = append(entries, kbDeliveryFailure)
	}
	if containsAny(lower, cancellationKeywordsKB) {
		entries = append(entries, kbCancellationAnomaly)
	}
	if containsAny(lower, refundKeywordsKB) {
		entries = append(entries, kbRefundAnomaly)
	}
	if containsAny(lower, stuckKeywordsKB) {
		entries = append(entries, kbStuckOrder)
	}

	// If note didn't match anything specific, include delivery + data_integrity
	// so the AI can still reason about generic anomalies from the driver note.
	if len(entries) == 1 {
		entries = append(entries, kbDeliveryFailure, kbDataIntegrity)
	}

	return entries
}

// containsAny returns true if s contains at least one of the given substrings.
func containsAny(s string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
