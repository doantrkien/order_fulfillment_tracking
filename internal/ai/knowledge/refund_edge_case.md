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