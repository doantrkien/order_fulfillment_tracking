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