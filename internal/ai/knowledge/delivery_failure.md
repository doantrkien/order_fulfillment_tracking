# Delivery Failure — Severity Rules

Exception type: DELIVERY_FAILURE

Applicable when: order is in 'shipped' status and the driver note describes a delivery problem.

## Severity decision

- **CRITICAL** — package lost, stolen, or unrecoverable (e.g. "lost", "stolen", "cannot be found")
- **HIGH** — physical obstacle prevents delivery: accident, vehicle breakdown, bad weather,
  could not deliver, giao thất bại, thời tiết, xe hỏng, tai nạn
- **MEDIUM** — access/contact issue: customer not home, no one home, wrong address,
  address not found, không có nhà, sai địa chỉ

## Internal next actions

- CRITICAL → Escalate to logistics manager. Open a lost-parcel investigation. Notify finance for potential claim.
- HIGH     → Escalate to logistics team. Arrange re-delivery or return-to-warehouse.
- MEDIUM   → Contact customer to reschedule delivery. Update delivery attempts log.

## Examples

Example 1 — CRITICAL (package lost):
Input: Driver note says "không tìm thấy kiện hàng, có thể đã bị mất". Order status: shipped.
Output:
{
  "exception_type": "DELIVERY_FAILURE",
  "severity": "CRITICAL",
  "likely_reason": "Package could not be located — potential loss in transit",
  "internal_next_action": "Escalate to logistics manager. Open a lost-parcel investigation. Notify finance for potential claim.",
  "confidence_score": 0.92
}

Example 2 — HIGH (vehicle breakdown):
Input: Driver note says "xe hỏng giữa đường, không thể giao hàng hôm nay". Order status: shipped.
Output:
{
  "exception_type": "DELIVERY_FAILURE",
  "severity": "HIGH",
  "likely_reason": "Delivery failed: vehicle breakdown prevented delivery",
  "internal_next_action": "Escalate to logistics team. Arrange re-delivery or return-to-warehouse.",
  "confidence_score": 0.93
}

Example 3 — MEDIUM (customer not home):
Input: Driver note says "đến địa chỉ nhưng không có ai ở nhà, đã gọi không nghe máy". Order status: shipped.
Output:
{
  "exception_type": "DELIVERY_FAILURE",
  "severity": "MEDIUM",
  "likely_reason": "Delivery failed: customer not home, no answer on phone",
  "internal_next_action": "Contact customer to reschedule delivery. Update delivery attempts log.",
  "confidence_score": 0.95
}
