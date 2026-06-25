# Alternative Delivery Success

Exception Type: OTHER

Applicable When:
The driver note indicates the package was successfully left in a safe/approved location (e.g., reception desk, security guard, neighbor, front door) because the customer was temporarily unreachable.

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
Log and monitor for future delivery attempts

# Example Output
Input
Status: delivered
Driver note: "Leave the package at the reception desk"
Output
{
  "exception_type": "OTHER",
  "severity": "LOW",
  "likely_reason": "Customer was unreachable but package was left at reception desk",
  "internal_next_action": "Log and monitor for future delivery attempts",
  "confidence_score": 0.95
}
