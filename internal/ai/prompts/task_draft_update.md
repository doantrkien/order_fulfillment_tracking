[TASK]
Draft a polite customer-facing message update regarding an order exception.
You must adopt the requested tone (e.g. apologetic, neutral, informative) and write appropriate length for the channel (e.g. SMS, Email).
Reassure the customer that we are resolving the issue.

[CONSTRAINTS]

1. NO INTERNAL JARGON: Do not leak internal system database states or raw error messages.
2. NO UNSUPPORTED PROMISES: Do not promise specific refund amounts or exact delivery times unless explicitly authorized.
3. PERSONAL DATA MASKING: You MUST use the static placeholders [REDACTED_CUSTOMER_NAME] and [REDACTED_SHIPPING_ADDRESS] to represent the customer's name and shipping address. Never output actual customer names or addresses.

[CONTEXT]
Order ID: {{ORDER_ID}}
Customer Name: [REDACTED_CUSTOMER_NAME]
Shipping Address: [REDACTED_SHIPPING_ADDRESS]
Current Status: {{CURRENT_STATUS}}
Exception Type: {{EXCEPTION_TYPE}}
Likely Reason: {{LIKELY_REASON}}
Requested Tone: {{TONE}}
Communication Channel: {{CHANNEL}}

[OUTPUT FORMAT]
Respond with EXACTLY this JSON structure:
{
"customer_update_draft": "<string: the drafted message in Vietnamese, with [REDACTED_CUSTOMER_NAME] and [REDACTED_SHIPPING_ADDRESS] left intact>",
"confidence_score": <float: 0.0 to 1.0, your confidence in the appropriateness of this draft>
}
