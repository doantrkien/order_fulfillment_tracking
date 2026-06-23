# Cancellation Anomaly — Severity Rules

Exception type: CANCELLATION_ANOMALY

Applicable when: order reaches 'cancelled' status after having been at a late stage.

## Severity decision (based on the status at the time of cancellation)

- **CRITICAL** — cancelled after 'shipped' or 'delivered'
- **HIGH**     — cancelled after 'packed'
- **MEDIUM**   — cancelled after 'paid' (early, less impactful)

## Internal next actions

- CRITICAL → Halt any ongoing delivery. Initiate return-to-warehouse procedure. Review refund eligibility.
- HIGH     → Coordinate with warehouse to stop packing. Confirm cancellation with customer.
- MEDIUM   → Process cancellation normally. Confirm with customer.

## Examples

Example 1 — CRITICAL (cancelled after shipped):
Input: Driver note says "khách hủy, đơn đang trên xe giao". Event timeline shows shipped → cancelled.
Output:
{
  "exception_type": "CANCELLATION_ANOMALY",
  "severity": "CRITICAL",
  "likely_reason": "Order was cancelled after reaching 'shipped' status — late-stage cancellation detected",
  "internal_next_action": "Halt any ongoing delivery. Initiate return-to-warehouse procedure. Review refund eligibility.",
  "confidence_score": 0.95
}

Example 2 — HIGH (cancelled after packed):
Input: Driver note says "khách không muốn nhận hàng nữa". Event timeline shows packed → cancelled.
Output:
{
  "exception_type": "CANCELLATION_ANOMALY",
  "severity": "HIGH",
  "likely_reason": "Order was cancelled after reaching 'packed' status",
  "internal_next_action": "Coordinate with warehouse to stop packing. Confirm cancellation with customer.",
  "confidence_score": 0.9
}

Example 3 — MEDIUM (cancelled after paid):
Input: Driver note says "customer requested cancellation". Event timeline shows paid → cancelled.
Output:
{
  "exception_type": "CANCELLATION_ANOMALY",
  "severity": "MEDIUM",
  "likely_reason": "Order was cancelled after payment — early-stage cancellation",
  "internal_next_action": "Process cancellation normally. Confirm with customer.",
  "confidence_score": 0.9
}
