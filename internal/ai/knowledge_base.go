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

//go:embed knowledge/alternative_success.md
var kbBodyAlternativeSuccess string

//go:embed knowledge/cancellation_edge_case.md
var kbBodyCancellation string

//go:embed knowledge/refund_edge_case.md
var kbBodyRefund string

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

var kbAlternativeSuccess = KnowledgeEntry{
	Title: "Alternative Delivery Success",
	Body:  kbBodyAlternativeSuccess,
}

var kbCancellation = KnowledgeEntry{
	Title: "Cancellation Edge Case",
	Body:  kbBodyCancellation,
}

var kbRefund = KnowledgeEntry{
	Title: "Refund Edge Case",
	Body:  kbBodyRefund,
}

var deliveryFailureKeywordsKB = []string{
	"lost", "stolen", "cannot find", "missing parcel", "hàng bị mất", "nghi thất lạc",

	"accident", "vehicle breakdown", "xe hỏng", "tai nạn", "bad weather", "thời tiết",
	"failed", "failure", "cannot deliver", "could not deliver", "giao thất bại",

	"not home", "no one home", "customer not home", "wrong address", "address not found",
	"không có nhà", "sai địa chỉ", "không liên lạc được",

	"temporarily unreachable", "no answer", "will retry", "khách không nghe máy",
}

var stateMachineKeywordsKB = []string{
	"tự động đổi trạng thái", "sai trạng thái hiện tại",
}

var duplicateKeywordsKB = []string{
	"bấm lại", "duplicate", "lặp", "trùng", "repeated", "nhiều lần", "times", "hệ thống ghi nhận delivered 2 lần.",
}

var skippedKeywordsKB = []string{
	"skipped", "missed", "bỏ qua", "thiếu", "không đi qua các bước thông thường",
}

var stuckKeywordsKB = []string{
	"stuck", "no progress", "not moved", "delay", "delayed", "late",
	"trễ", "chậm", "không tiến triển", "Đơn hàng bị giữ lại",
}

var successKeywordsKB = []string{
	"reception", "lễ tân", "neighbor", "hàng xóm", "bảo vệ", "security", "front door", "trước cửa", "thành công", "delivered",
}

var cancellationKeywordsKB = []string{
	"cancel",
	"cancelled",
	"cancellation",
	"customer requested cancellation",
	"cancel request",
	"hủy",
	"hủy đơn",
	"yêu cầu hủy",
	"đổi ý",
}

var refundKeywordsKB = []string{
	"refund",
	"refunded",
	"refund request",
	"duplicate payment",
	"chargeback",
	"hoàn tiền",
	"yêu cầu hoàn tiền",
	"thanh toán nhầm",
}

func ClassifyDriverNote(note string) []KnowledgeEntry {
	lower := strings.ToLower(strings.TrimSpace(note))
	entries := []KnowledgeEntry{} // always present

	matched := false
	isSuccessAlternative := containsAny(lower, successKeywordsKB)

	if isSuccessAlternative {
		entries = append(entries, kbAlternativeSuccess)
		matched = true
	}

	if containsAny(lower, deliveryFailureKeywordsKB) && !isSuccessAlternative {
		entries = append(entries, kbDeliveryFailure)
		matched = true
	}

	if containsAny(lower, stateMachineKeywordsKB) && !isSuccessAlternative {
		entries = append(entries, kbStateMachine)
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

	if containsAny(lower, cancellationKeywordsKB) {
		entries = append(entries, kbCancellation)
		matched = true
	}

	if containsAny(lower, refundKeywordsKB) {
		entries = append(entries, kbRefund)
		matched = true
	}

	if !matched {
		entries = append(entries, kbDeliveryFailure, kbDuplicateEvent, kbSkippedStatus, kbStuckOrder, kbCancellation, kbRefund)
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
