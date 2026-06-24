# Duplicate Event — Severity Rules

Exception Type: DUPLICATE_EVENT

Applicable When:
The same event (same order + same status) is received more than once.

## Detection Rule
Duplicate if (order_id + status) already exists in event history and is received again


## Severity Classification
### LOW
Condition: No side effect, fully idempotent

### MEDIUM
Condition: Causes redundant internal processing only
paid → paid → triggers reprocessing pipeline

### HIGH
Condition: Causes external side effects (notifications, logs, downstream calls)
shipped → shipped → duplicate shipping notification sent

### CRITICAL

Condition: Causes incorrect business or financial state
delivered → delivered → refund triggered twice

## Example Outputs
### Example 1 — LOW
Input
Event timeline:
Output
{
  "exception_type": "DUPLICATE_EVENT",
  "severity": "LOW",
  "likely_reason": "Duplicate event detected: status 'paid' received more than once with no side effect",
  "internal_next_action": "Log and monitor event source for retry behavior",
  "confidence_score": 0.98
}

### Example 2 — MEDIUM

Input

Event timeline:
Output
{
  "exception_type": "DUPLICATE_EVENT",
  "severity": "MEDIUM",
  "likely_reason": "Duplicate event triggered unnecessary internal reprocessing for status 'paid'",
  "internal_next_action": "Investigate consumer idempotency and reduce redundant processing",
  "confidence_score": 0.96
}

### Example 3 — HIGH
Input
Event timeline:
Output
{
  "exception_type": "DUPLICATE_EVENT",
  "severity": "HIGH",
  "likely_reason": "Duplicate event caused repeated external side effects (shipping notification)",
  "internal_next_action": "Fix idempotency in downstream services and add deduplication layer",
  "confidence_score": 0.97
}

### Example 4 — CRITICAL
Input
Event timeline:
delivered → delivered (refund triggered twice)
Output
{
  "exception_type": "DUPLICATE_EVENT",
  "severity": "CRITICAL",
  "likely_reason": "Duplicate event caused incorrect financial operation: refund executed twice",
  "internal_next_action": "Stop processing, rollback incorrect transactions, and audit event pipeline",
  "confidence_score": 0.99
}

