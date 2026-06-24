# Invalid Transition — Severity Rules

Exception Type: INVALID_TRANSITION

Applicable When:
An order transitions from one status to another status that is not allowed by the defined state machine.

# Valid Order Statuses
created
paid
packed
shipped
delivered
cancelled
refunded

# Valid State Transitions
created  -> paid, cancelled
paid     -> packed, refunded
packed   -> shipped
shipped  -> delivered

# Severity Classification
LOW
Condition
A transition event is received but the destination status is identical to the current status.
## Examples:
created -> created
paid -> paid
shipped -> shipped
Description
Likely caused by duplicate events or retry mechanisms.
Recommended Action
Log and monitor for duplicate processing.

# MEDIUM
Condition
The transition skips exactly one expected step in the workflow.
## Examples:
created -> packed
paid -> shipped
packed -> delivered
Description
Possible synchronization issue or missing event.
Recommended Action
Investigate event ordering and processing logic.

# HIGH
Condition
The transition skips multiple workflow stages.
## Examples:
created -> shipped
created -> delivered
paid -> delivered
Description
Strong indication of workflow corruption or service integration failure.
Recommended Action
Escalate to the responsible engineering team.

# CRITICAL
Condition
Any of the following:
## Transition originates from a terminal state
delivered -> shipped
cancelled -> paid
refunded -> packed
## Reverse transition
packed -> paid
shipped -> packed
delivered -> shipped
## Transition to an unrelated state that violates business rules
created -> refunded
packed -> cancelled
shipped -> refunded
Description
The order state machine has been violated. Data integrity may be compromised.
Recommended Action
Immediately block processing, investigate root cause, and perform data consistency checks.

# Example Outputs
Example 1 — LOW
Input
Current status: paid
New status: paid
Output
{
  "exception_type": "INVALID_TRANSITION",
  "severity": "LOW",
  "likely_reason": "Duplicate transition event received for status 'paid'",
  "internal_next_action": "Log duplicate event and monitor retry behavior",
  "confidence_score": 0.95
}

# Example 2 — MEDIUM
Input
Current status: created
New status: packed
Output
{
  "exception_type": "INVALID_TRANSITION",
  "severity": "MEDIUM",
  "likely_reason": "Transition skipped expected status 'paid'",
  "internal_next_action": "Investigate event ordering and synchronization issues",
  "confidence_score": 0.93
}

# Example 3 — HIGH
Input
Current status: created
New status: delivered
Output
{
  "exception_type": "INVALID_TRANSITION",
  "severity": "HIGH",
  "likely_reason": "Transition skipped multiple workflow stages",
  "internal_next_action": "Escalate to engineering team and investigate workflow integrity",
  "confidence_score": 0.97
}

# Example 4 — CRITICAL
Input
Current status: delivered
New status: shipped
Output
{
  "exception_type": "INVALID_TRANSITION",
  "severity": "CRITICAL",
  "likely_reason": "Attempted transition from terminal state 'delivered'",
  "internal_next_action": "Block processing immediately and perform data consistency checks",
  "confidence_score": 0.99
}
