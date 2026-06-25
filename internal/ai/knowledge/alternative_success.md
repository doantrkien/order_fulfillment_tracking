# Alternative Delivery Success

Exception Type: OTHER

Applicable When:
The driver note indicates the package was successfully left in a safe/approved location (e.g., reception desk, security guard, neighbor, front door) either per the customer's request or because the customer was temporarily unreachable.

This is a SUCCESSFUL delivery, NOT a failure.

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
No action required. Order is fulfilled successfully.

# Example Output
Input
Status: delivered
Driver note: "Leave the package at the reception desk"
Output
{
  "exception_type": "OTHER",
  "severity": "LOW",
  "likely_reason": "[Summarize the driver note comprehensively. State exactly whether the customer was contacted (if mentioned) and the final package location. DO NOT hallucinate that the customer was unreachable if the note says they were contacted. Do not copy this placeholder.]",
  "internal_next_action": "No action required. Order is fulfilled successfully.",
  "confidence_score": 0.95
}
