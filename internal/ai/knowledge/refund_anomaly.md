# Refund Anomaly — Severity Rules

Exception type: REFUND_ANOMALY

Applicable when: order reaches 'refunded' status in a suspicious way.

## Severity decision

- **CRITICAL** — refunded from 'created' status (order was never paid — potential fraud)
- **HIGH**     — refunded directly from 'paid' status (bypassed normal cancellation flow,
  risk of double-refund)

## Internal next actions

- CRITICAL → Immediately investigate payment gateway logs. Reverse the refund transaction if fraudulent. Escalate to finance.
- HIGH     → Verify refund legitimacy with payment team. Check for duplicate refund requests. Audit payment gateway records.

## Examples

Example 1 — CRITICAL (refunded from 'created', never paid):
Input: Driver note says "khách yêu cầu hoàn tiền nhưng chưa thanh toán bao giờ". Event timeline shows created → refunded.
Output:
{
  "exception_type": "REFUND_ANOMALY",
  "severity": "CRITICAL",
  "likely_reason": "Refund was processed for an order that was never paid — potential fraud",
  "internal_next_action": "Immediately investigate payment gateway logs. Reverse the refund transaction if fraudulent. Escalate to finance.",
  "confidence_score": 0.97
}

Example 2 — HIGH (refunded directly from 'paid', skipped cancellation):
Input: Driver note says "hoàn tiền trực tiếp theo yêu cầu khẩn cấp". Event timeline shows paid → refunded (no cancelled step).
Output:
{
  "exception_type": "REFUND_ANOMALY",
  "severity": "HIGH",
  "likely_reason": "Refund processed directly from paid status — bypassed normal cancellation flow, risk of double-refund",
  "internal_next_action": "Verify refund legitimacy with payment team. Check for duplicate refund requests. Audit payment gateway records.",
  "confidence_score": 0.91
}
