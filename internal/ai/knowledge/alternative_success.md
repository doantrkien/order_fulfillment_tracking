# Alternative Delivery Success

Exception Type: ALTERNATIVE_DELIVERY

Applicable When:
The driver note indicates the package was successfully left in a safe/approved location (e.g., reception desk, security guard, neighbor, front door) either per the customer's request or because the customer was temporarily unreachable.

This is a SUCCESSFUL delivery, NOT a failure. However, it deviates from the standard happy-path delivery (direct handoff to the customer) and must be recorded for audit, risk mitigation, and driver KPI evaluation.

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
