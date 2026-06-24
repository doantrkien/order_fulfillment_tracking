# Delivery Failure — Severity Rules

Exception Type: DELIVERY_FAILURE

Applicable When:
Order is in shipped status and a delivery attempt fails due to a driver-reported issue.

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