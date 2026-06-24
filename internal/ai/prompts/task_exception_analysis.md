[DOMAIN KNOWLEDGE]
Valid order statuses: created, paid, packed, shipped, delivered, cancelled, refunded
Valid state transitions:
created → paid, cancelled
paid → packed, refunded
packed → shipped
shipped → delivered
Terminal states (no further transitions): delivered, cancelled, refunded

{{KNOWLEDGE_BASE_INJECTION}}

[TASK]
Analyze the provided order context and identify any exception or anomaly. Consider:

- Invalid or unexpected status transitions
- Stuck orders (no progress for an unusually long time)
- Skipped statuses in the lifecycle
- Duplicate or contradictory events
- Delivery failures or cancellations with unusual patterns
- Any anomaly mentioned in the operator notes

[CONTEXT]
Order ID: {{ORDER_ID}}
Current Status: {{CURRENT_STATUS}}
Total Amount: {{TOTAL_AMOUNT}} VND
Order Created At: {{CREATED_AT}}
Analysis Timestamp (now): {{ANALYZED_AT}}

Event Timeline (chronological order):
{{EVENT_TIMELINE}}

Operator Notes:
{{DRIVER_NOTES}}

[OUTPUT FORMAT]
Respond with EXACTLY this JSON structure:
{
"exception_type": "<string: one of INVALID_TRANSITION | STUCK_ORDER | SKIPPED_STATUS | DUPLICATE_EVENT | DELIVERY_FAILURE | CANCELLATION_ANOMALY | REFUND_ANOMALY | OTHER>",
"severity": "<string: one of LOW | MEDIUM | HIGH | CRITICAL>",
"likely_reason": "<string: concise root cause explanation in English, max 200 chars>",
"internal_next_action": "<string: recommended internal action for the fulfillment team, max 200 chars>",
"suggestion": "<string: additional detailed suggestions or context, max 300 chars>",
"should_alert": <boolean: true if the exception warrants an alert, false otherwise>,
"confidence_score": <float: 0.0 to 1.0, your confidence in this analysis>
}
