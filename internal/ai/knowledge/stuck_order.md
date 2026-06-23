# Stuck Order — Severity Rules

Exception type: STUCK_ORDER

Applicable when: order has not advanced from its current status for longer than the normal threshold.

## Normal thresholds by status

| Status   | Threshold | Severity baseline |
|----------|-----------|-------------------|
| created  | 24 h      | MEDIUM            |
| paid     | 48 h      | MEDIUM            |
| packed   | 24 h      | HIGH              |
| shipped  | 72 h      | HIGH              |

## Severity decision

- **CRITICAL** — stuck time exceeds 2× the normal threshold for that status
- **HIGH**     — stuck in 'packed' > 24h or 'shipped' > 72h (not yet 2×)
- **MEDIUM**   — stuck in 'created' > 24h or 'paid' > 48h (not yet 2×)

## Internal next action

Investigate why the order has not advanced. Contact the responsible team.

## Examples

Example 1 — CRITICAL (stuck 2× threshold, 'shipped' > 144h):
Input: Driver note says "đơn bị trễ, chưa cập nhật gì thêm". Order in 'shipped' status for 145 hours (threshold: 72h, 2×=144h).
Output:
{
  "exception_type": "STUCK_ORDER",
  "severity": "CRITICAL",
  "likely_reason": "Order has been in 'shipped' status for 145 hours — exceeds 2× the 72h threshold",
  "internal_next_action": "Escalate urgently. Investigate why order has not advanced from 'shipped'. Contact logistics team immediately.",
  "confidence_score": 0.95
}

Example 2 — HIGH (stuck in 'packed' for 30h, under 2×):
Input: Driver note says "chưa lấy hàng". Order in 'packed' status for 30 hours (threshold: 24h).
Output:
{
  "exception_type": "STUCK_ORDER",
  "severity": "HIGH",
  "likely_reason": "Order has been in 'packed' status for 30 hours without progressing",
  "internal_next_action": "Investigate why order has not advanced from 'packed'. Contact the responsible team.",
  "confidence_score": 0.9
}

Example 3 — MEDIUM (stuck in 'paid' for 50h, under 2×):
Input: Driver note says "chưa có ai đến lấy hàng". Order in 'paid' status for 50 hours (threshold: 48h).
Output:
{
  "exception_type": "STUCK_ORDER",
  "severity": "MEDIUM",
  "likely_reason": "Order has been in 'paid' status for 50 hours without progressing",
  "internal_next_action": "Investigate why order has not advanced from 'paid'. Contact the warehouse team.",
  "confidence_score": 0.88
}
