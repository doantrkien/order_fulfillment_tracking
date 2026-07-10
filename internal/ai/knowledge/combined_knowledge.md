# alternative_success.md
# Alternative Delivery Success

Exception Type: ALTERNATIVE_DELIVERY

Applicable When:
The driver note indicates the package was successfully left in a safe/approved location (e.g., reception desk, security guard, neighbor, front door) either per the customer's request or because the customer was temporarily unreachable.

This is a SUCCESSFUL delivery, NOT a failure. However, it deviates from the standard happy-path delivery (direct handoff to the customer) and must be recorded for audit, risk mitigation, and driver KPI evaluation. A successful standard delivery or repeated confirmation of delivery is NOT an ALTERNATIVE_DELIVERY exception.

## Severity Classification
### LOW
Condition
The package is safe and delivered to an authorized alternative receiver.
Examples:
Leave the package at the reception desk
Gửi bảo vệ rồi
Giao cho hàng xóm
Likely Impact
Order is completed successfully.
Internal Next Action
Generate a specific, contextual action based on the actual driver note. Do NOT copy the example below verbatim. Consider the specific delivery location, whether the customer was contacted, and any follow-up needed.

# Example Output
Input
Status: delivered
Driver note: "Leave the package at the reception desk"
Output
{
  "exception_type": "ALTERNATIVE_DELIVERY",
  "severity": "LOW",
  "likely_reason": "The package was successfully delivered and left at the reception desk for the customer to pick up later.",
  "internal_next_action": "Record alternative handoff location (reception desk). Flag for customer confirmation if no pickup within 24h.",
  "confidence_score": 0.95
}

---
# cancellation_edge_case.md
# Cancellation Edge Case

## Definition

A cancellation edge case occurs when a cancellation request is made at different stages of the order lifecycle. The request must be validated against the current order status.

---

## Severity Classification

### LOW

Condition

Customer cancels immediately after order creation.

Example

created → cancelled

Description

This is a normal business scenario.

Recommended Action

Process the cancellation normally.

---

### MEDIUM

Condition

Customer requests cancellation after payment but before packing.

Example

created → paid → cancelled

Description

The cancellation is generally allowed but payment should be verified.

Recommended Action

Verify payment and initiate the refund process if required.

---

### HIGH

Condition

Customer requests cancellation after packing has started.

Example

created → paid → packed → cancelled

Description

Warehouse operations may already be in progress.

Recommended Action

Confirm whether packing can still be stopped before approving cancellation.

---

### CRITICAL

Condition

Cancellation occurs after shipment has started or after delivery.

Example

created → paid → packed → shipped → cancelled

created → paid → packed → shipped → delivered → cancelled

Description

This is an invalid workflow and may indicate incorrect events or data inconsistency.

Recommended Action

Reject the cancellation request and investigate the order timeline.

---

## Additional Examples

Example 1

Current Status:
packed

Driver Note:
Customer changed their mind before shipment.

Expected Result:
Severity: HIGH

---

Example 2

Current Status:
shipped

Driver Note:
Customer requested cancellation after receiving the tracking number.

Expected Result:
Severity: CRITICAL

---
# delivery_failure.md
# Delivery Failure — Severity Rules

Exception Type: DELIVERY_FAILURE

Applicable When:
Order is in shipped status and a delivery attempt fails due to a driver-reported issue.

Exceptions (Alternative Delivery Success):
If the driver note indicates the customer was unreachable BUT the package was successfully left in a safe/approved location (e.g., reception desk, security guard, neighbor, front door), this is a SUCCESSFUL alternative delivery.
DO NOT classify as DELIVERY_FAILURE. Classify as OTHER with LOW severity, and recommend monitoring or no action.

## Severity Classification
### LOW — Temporary / Minor Issue
Condition
Minor issue that does not block delivery and can be resolved immediately or during the same attempt.
Examples:
customer temporarily unreachable
no answer, will retry call
short delay at delivery point
khách không nghe máy tạm thời
Likely Impact
Delivery can still succeed in the same day or next immediate attempt.
Internal Next Action
Retry contact customer and reattempt delivery.

### MEDIUM — Customer or Address Issue (Retry Required)
Condition
Delivery failed due to customer availability or address-related issues requiring rescheduling.
Examples:
customer not home
no one available to receive package
wrong address
address not found
không có ai ở nhà
sai địa chỉ
không liên lạc được khách hàng
Likely Impact
Delivery can be completed after rescheduling.
Internal Next Action
Contact customer to reschedule delivery and log attempt.

### HIGH — Operational / External Disruption
Condition
Delivery cannot be completed due to external or operational obstacles.
Examples:
vehicle breakdown
accident on route
bad weather
road blocked
xe hỏng
tai nạn giao thông
thời tiết xấu
Likely Impact
Delivery delayed but still recoverable via re-delivery or return-to-warehouse.
Internal Next Action
Escalate to logistics team. Arrange re-delivery or return-to-warehouse.

### CRITICAL — Lost or Unrecoverable Package
Condition
Package is lost, stolen, or cannot be located.
Examples:
package lost
stolen
cannot find package
missing parcel
hàng bị mất
không tìm thấy kiện hàng
nghi thất lạc
Likely Impact
Order cannot be fulfilled without investigation or compensation.
Internal Next Action
Escalate to logistics manager. Open lost-parcel investigation and notify finance for claim handling.

## Severity Priority Rule
## Examples
### Example 1 — LOW
Input
Status: `shipped`
Driver note: `"khách đang bận, sẽ gọi lại sau"`
Output
{
  "exception_type": "DELIVERY_FAILURE",
  "severity": "LOW",
  "likely_reason": "Customer temporarily unreachable but delivery can be retried immediately",
  "internal_next_action": "Retry contact customer and reattempt delivery",
  "confidence_score": 0.90
}

### Example 2 — MEDIUM
Input
Status: `shipped`
Driver note: `"không có ai ở nhà, gọi không nghe máy"`
Output
{
  "exception_type": "DELIVERY_FAILURE",
  "severity": "MEDIUM",
  "likely_reason": "Customer not available at delivery location",
  "internal_next_action": "Contact customer to reschedule delivery and log attempt",
  "confidence_score": 0.95
}

### Example 3 — HIGH
Input
Status: `shipped`
Driver note: `"xe hỏng giữa đường, không thể giao"`
Output
{
  "exception_type": "DELIVERY_FAILURE",
  "severity": "HIGH",
  "likely_reason": "Operational issue (vehicle breakdown) prevented delivery",
  "internal_next_action": "Escalate to logistics team. Arrange re-delivery or return-to-warehouse.",
  "confidence_score": 0.94
}


### Example 4 — CRITICAL
Input
Status: `shipped`
Driver note: `"không tìm thấy kiện hàng, nghi bị mất"`
Output
{
  "exception_type": "DELIVERY_FAILURE",
  "severity": "CRITICAL",
  "likely_reason": "Package is missing or potentially lost in transit",
  "internal_next_action": "Escalate to logistics manager. Open lost-parcel investigation and notify finance for claim handling.",
  "confidence_score": 0.92
}

---
# duplicate_event.md
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

---
# prompt_draft.md
[SYSTEM]
You are a Customer Support Agent for a fulfillment tracking system.
Your job is to draft a customer-facing update message regarding an order exception.
You MUST avoid making unsupported promises (like guaranteed delivery times) and MUST NOT leak internal technical details.
You MUST respond ONLY with a single valid JSON object. No explanations, no markdown, no extra text.

[CONTEXT]
%s

[BASELINE DRAFT]
This is the standard company template for this exception:
%s

[TASK]
Draft a message to the customer explaining the situation politely.
Use the [BASELINE DRAFT] as your foundation. You must modify it to match the requested Tone and Channel length.
If the exception is OTHER, use the Likely Reason to provide more specific details.
Reassure them that we are handling it, but do not promise refunds or exact resolution times unless explicitly supported by standard policy.

[OUTPUT FORMAT]
Respond with EXACTLY this JSON structure:
{
  "customer_update_draft": "<string: the drafted message for the customer>",
  "confidence_score": <float: 0.0 to 1.0, your confidence in the appropriateness of this draft>
}

[CONSTRAINTS]
1. NO internal technical jargon.
2. NO false promises.
3. DO NOT output anything other than the JSON object.

---
# prompt_exception.md
[SYSTEM]
You are an Order Exception Analyst for a fulfillment tracking system.
Your role is to analyze order data, event history, and operator notes to identify exceptions.
You must determine the exception type, severity, likely root cause, and recommend the next internal action.
Use ONLY the rules provided in [KNOWLEDGE BASE] to assign severity — do not use intuition or guesswork.
You MUST respond ONLY with a single valid JSON object. No explanations, no markdown, no extra text.

[CONTEXT]
%s

[KNOWLEDGE BASE]
Apply the following domain rules when classifying and rating the exception:

%s

[TASK]
Analyze the order context and the driver note above.
Identify the exception type and assign severity using ONLY the [KNOWLEDGE BASE] rules.

[OUTPUT FORMAT]
Respond with EXACTLY this JSON structure (no additional fields, no wrapping):
{
  "exception_type": "<string: one of INVALID_TRANSITION | STUCK_ORDER | SKIPPED_STATUS | DUPLICATE_EVENT | DELIVERY_FAILURE | ALTERNATIVE_DELIVERY | CANCELLATION_ANOMALY | REFUND_ANOMALY | OTHER>",
  "severity": "<string: one of LOW | MEDIUM | HIGH | CRITICAL>",
  "likely_reason": "<string: concise root cause explanation in English, max 200 chars>",
  "internal_next_action": "<string: recommended internal action for the fulfillment team, max 200 chars>",
  "should_alert": <boolean: true if the exception warrants an alert, false otherwise>,
  "confidence_score": <float: 0.0 to 1.0, your confidence in this analysis>
}

[CONSTRAINTS]
1. You MUST NOT suggest or imply any automatic order status changes.
2. You MUST NOT suggest or trigger any refund actions.
3. You MUST NOT suggest sending any messages or notifications to customers.
4. Your role is ANALYSIS ONLY — observe, diagnose, and recommend internal actions.
5. If you cannot determine the exception with reasonable confidence, set confidence_score below 0.5.
6. Do NOT output anything other than the JSON object. No markdown fences, no explanations.
7. If the driver note is clearly nonsense, irrelevant, profane, or a joke, classify as "OTHER" with severity "LOW" and indicate that the note is unhelpful or inappropriate.
8. Severity MUST be determined by evaluating both the order lifecycle history and the content of the driver note. A critical note (e.g. accident, lost package) elevates severity regardless of timeline.
9. If the order lifecycle shows that one or more mandatory workflow stages were skipped, classify as "SKIPPED_STATUS" (unless a more severe exception applies).
10. If the provided [KNOWLEDGE BASE] does NOT contain rules that match the situation or you are uncertain, you MUST classify it as "OTHER" rather than forcing it into an unrelated category.

---
# refund_edge_case.md
# Refund Edge Case

## Definition

A refund edge case occurs when a refund request is received at different stages of the order lifecycle.

---

## Severity Classification

### LOW

Condition

Refund requested immediately after payment.

Example

created → paid → refunded

Description

A normal refund scenario.

Recommended Action

Process the refund according to payment policy.

---

### MEDIUM

Condition

Refund requested after packing has begun.

Example

created → paid → packed → refunded

Description

Inventory allocation should be reviewed.

Recommended Action

Coordinate with warehouse staff before approving the refund.

---

### HIGH

Condition

Refund requested after shipment has started.

Example

created → paid → packed → shipped → refunded

Description

The shipment may already be in transit.

Recommended Action

Verify whether shipment can be intercepted or returned.

---

### CRITICAL

Condition

Refund recorded after successful delivery without any return process.

Example

created → paid → packed → shipped → delivered → refunded

Description

This may indicate fraud, duplicate processing, or inconsistent order events.

Recommended Action

Review payment records, return records, and the complete order history before approving the refund.

---

## Additional Examples

Example 1

Current Status:
paid

Driver Note:
Customer accidentally placed the order twice and requested a refund.

Expected Result:
Severity: LOW

---

Example 2

Current Status:
shipped

Driver Note:
Customer requested a refund while the parcel is already in transit.

Expected Result:
Severity: HIGH

---

Example 3

Current Status:
delivered

Driver Note:
A refund was processed even though no return request exists.

Expected Result:
Severity: CRITICAL

---
# skipped_status.md
# Skipped Status — Severity Rules

Exception Type: SKIPPED_STATUS

Applicable When:
One or more mandatory statuses are missing from the order lifecycle history.

## Order Lifecycle
created -> paid -> packed -> shipped -> delivered

Alternative terminal paths:
created -> cancelled
paid -> refunded

## Severity Classification
### LOW
Condition
Exactly one status is skipped near the beginning of the workflow.
Examples:
created -> packed
Skipped:
paid
Description
A single intermediate status is missing, so event consistency should be verified.
Recommended Action
Review event logs and monitor for additional anomalies.

### MEDIUM
Condition
Exactly one status is skipped in a fulfillment or logistics stage.
Examples:
paid -> shipped
Skipped:
packed
Description
A required operational step appears to be missing. This may indicate event loss or process bypass.
Recommended Action
Investigate the order processing pipeline and verify status generation.

### HIGH
Condition
Two consecutive required statuses are skipped.
Examples:
created -> shipped
Skipped:
paid, packed
paid -> delivered
Skipped:
packed, shipped
Description
Multiple mandatory workflow steps are missing. This strongly suggests a processing or integration issue.
Recommended Action
Escalate to the responsible engineering or operations team.

### CRITICAL
Condition
Three or more required statuses are skipped.
Examples:
created -> delivered
Skipped:
paid, packed, shipped
created -> refunded
Skipped:
paid
and violates the expected business flow.
Description
A significant portion of the lifecycle is missing. Data integrity may be compromised.
Recommended Action
Immediately investigate the order history, validate event consistency, and perform a full audit of the processing pipeline.

### Example Outputs
LOW
{
  "exception_type": "SKIPPED_STATUS",
  "severity": "LOW",
  "likely_reason": "Order transitioned from 'created' directly to 'packed', skipping required status 'paid'",
  "internal_next_action": "Review event logs and monitor for additional anomalies",
  "confidence_score": 0.88
}
MEDIUM
{
  "exception_type": "SKIPPED_STATUS",
  "severity": "MEDIUM",
  "likely_reason": "Order transitioned from 'paid' directly to 'shipped', skipping required status 'packed'",
  "internal_next_action": "Investigate the order processing pipeline and verify status generation",
  "confidence_score": 0.92
}
HIGH
{
  "exception_type": "SKIPPED_STATUS",
  "severity": "HIGH",
  "likely_reason": "Order skipped multiple required statuses: 'paid' and 'packed'",
  "internal_next_action": "Escalate to the responsible engineering team and investigate workflow integrity",
  "confidence_score": 0.95
}
CRITICAL
{
  "exception_type": "SKIPPED_STATUS",
  "severity": "CRITICAL",
  "likely_reason": "Order transitioned from 'created' directly to 'delivered', skipping several mandatory lifecycle stages",
  "internal_next_action": "Perform immediate data integrity audit and investigate the processing pipeline",
  "confidence_score": 0.99
}

---
# state_machine.md
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

---
# stuck_order.md
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

For STUCK_ORDER analysis:
- Treat analysis_time as the current system time.
- Calculate the time elapsed since the most recent event in event_history.
- Compare it against the configured threshold to determine severity.

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
