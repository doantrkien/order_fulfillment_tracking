# Invalid Transition — Severity Rules

Exception Type: INVALID_TRANSITION

Applicable When:

An order transitions to a status that is impossible according to the defined state machine, excluding cases where the transition can be explained solely by skipped intermediate statuses.

If the primary anomaly is one or more missing lifecycle statuses, classify it as **SKIPPED_STATUS** instead.

# Valid Order Statuses
created
paid
packed
shipped
delivered
cancelled
refunded

# Valid State Transitions
created → paid, cancelled
paid → packed, refunded
packed → shipped
shipped → delivered

# Severity Classification
## MEDIUM
Condition
A transition violates the state machine but does not originate from a terminal state and is not a reverse transition.
Examples
created → refunded
packed → cancelled
shipped → refunded
Description
The order entered a business state that is not reachable from the current status.
Recommended Action
Investigate workflow validation and service integration logic.

## HIGH
Condition
A reverse transition occurs.
Examples
packed → paid
shipped → packed
Description
The order moved backwards in the lifecycle, indicating possible event ordering problems or data inconsistency.
Recommended Action
Escalate to engineering and verify event ordering and state synchronization.

## CRITICAL
Condition
The transition originates from a terminal state.
Examples
delivered → shipped
delivered → packed
cancelled → paid
cancelled → shipped
refunded → packed
refunded → delivered
Description
A terminal order state was modified after completion. Data integrity may be compromised.
Recommended Action
Immediately block processing, investigate the root cause, and perform data consistency checks.

# Important Rule
Do NOT classify the following as INVALID_TRANSITION:
created → packed
paid → shipped
packed → delivered
created → shipped
created → delivered
paid → delivered
These cases represent missing mandatory lifecycle stages and MUST be classified as SKIPPED_STATUS.


# Example Outputs
## Example 1 — LOW
Input
Current status: paid
New status: paid
Output
{
"exception_type": "INVALID_TRANSITION",
"severity": "LOW",
"likely_reason": "Duplicate transition event received for status 'paid'",
"internal_next_action": "Log duplicate event and monitor retry behavior",
"confidence_score": 0.96
}

## Example 2 — MEDIUM
Input
Current status: packed
New status: cancelled
Output
{
"exception_type": "INVALID_TRANSITION",
"severity": "MEDIUM",
"likely_reason": "Transition from 'packed' to 'cancelled' is not allowed by the state machine",
"internal_next_action": "Investigate workflow validation and service integration",
"confidence_score": 0.95
}

## Example 3 — HIGH
Input
Current status: shipped
New status: packed
Output
{
"exception_type": "INVALID_TRANSITION",
"severity": "HIGH",
"likely_reason": "Reverse transition detected from 'shipped' to 'packed'",
"internal_next_action": "Escalate to engineering and verify event ordering",
"confidence_score": 0.98
}

## Example 4 — CRITICAL
Input
Current status: delivered
New status: shipped
Output
{
"exception_type": "INVALID_TRANSITION",
"severity": "CRITICAL",
"likely_reason": "Attempted transition from terminal state 'delivered'",
"internal_next_action": "Immediately block processing and perform data consistency checks",
"confidence_score": 0.99
}
