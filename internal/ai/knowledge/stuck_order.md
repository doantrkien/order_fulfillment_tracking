# Stuck Order — Severity Rules

Exception Type: STUCK_ORDER

Applicable When:
An order remains in the same status longer than the expected processing threshold and has not advanced to the next stage.

Normal Thresholds by Status
Status	Threshold
created	24 hours
paid	48 hours
packed	24 hours
shipped	72 hours

## Severity Classification

Severity is determined based on the ratio between the current stuck time and the configured threshold for the order's current status.

## LOW

Condition

threshold < stuck_time ≤ 1.25 × threshold

Description

The order is slightly delayed beyond the expected processing time. This may occur during normal operational fluctuations.

Recommended Action

Monitor the order and notify the responsible team.

## MEDIUM

Condition

1.25 × threshold < stuck_time ≤ 2 × threshold

Description

The order is significantly delayed and requires investigation to determine why it has not progressed.

Recommended Action

Investigate the delay and contact the responsible team.

## HIGH

Condition

2 × threshold < stuck_time ≤ 3 × threshold

Description

The order appears to be blocked, overlooked, or impacted by an operational issue.

Recommended Action

Escalate to the team lead and investigate immediately.

## CRITICAL

Condition

stuck_time > 3 × threshold

Description

The order is severely delayed and may have been abandoned, forgotten, or affected by a major operational failure.

Recommended Action

Perform urgent escalation and initiate immediate investigation.

Internal Next Actions
Severity	Internal Next Action
LOW	Monitor and notify the responsible team
MEDIUM	Investigate the delay and contact the responsible team
HIGH	Escalate to the team lead and investigate immediately
CRITICAL	Urgently escalate to operations management and investigate immediately

## Examples
### Example 1 — LOW

Input
Current status: paid
Time in status: 55 hours
Threshold: 48 hours

Output

{
  "exception_type": "STUCK_ORDER",
  "severity": "LOW",
  "likely_reason": "Order has been in 'paid' status for 55 hours, slightly exceeding the normal 48-hour threshold",
  "internal_next_action": "Monitor and notify the responsible team",
  "confidence_score": 0.85
}
### Example 2 — MEDIUM

Input
Current status: packed
Time in status: 40 hours
Threshold: 24 hours

Output

{
  "exception_type": "STUCK_ORDER",
  "severity": "MEDIUM",
  "likely_reason": "Order has been in 'packed' status for 40 hours without progressing",
  "internal_next_action": "Investigate why the order has not advanced from 'packed'. Contact the responsible team.",
  "confidence_score": 0.90
}

### Example 3 — HIGH

Input
Current status: shipped
Time in status: 170 hours
Threshold: 72 hours

Output

{
  "exception_type": "STUCK_ORDER",
  "severity": "HIGH",
  "likely_reason": "Order has been in 'shipped' status for 170 hours, exceeding 2× the normal threshold",
  "internal_next_action": "Escalate to the logistics team lead and investigate immediately",
  "confidence_score": 0.94
}

### Example 4 — CRITICAL

Input
Current status: shipped
Time in status: 230 hours
Threshold: 72 hours

Output

{
  "exception_type": "STUCK_ORDER",
  "severity": "CRITICAL",
  "likely_reason": "Order has been in 'shipped' status for 230 hours, exceeding 3× the normal threshold",
  "internal_next_action": "Urgent escalation to operations management. Investigate immediately.",
  "confidence_score": 0.97
}
