# Data Integrity Issues — Severity Rules

## SKIPPED_STATUS

Applicable when: a mandatory lifecycle step is missing from the event history
(e.g. order went from 'paid' directly to 'shipped', skipping 'packed').

- Severity: always **HIGH**
- Internal next action: Verify the order processing pipeline. Missing status may indicate
  a system bypass or data integrity issue.

## DUPLICATE_EVENT

Applicable when: the same status appears consecutively in the event log.

- Severity: always **LOW**
- Internal next action: Investigate the event source for duplicate submissions. No immediate action required.

## Examples

Example 1 — SKIPPED_STATUS (HIGH):
Input: Driver note says "hệ thống tự chuyển sang shipped mà không qua packed". Event timeline shows paid → shipped (packed is missing).
Output:
{
  "exception_type": "SKIPPED_STATUS",
  "severity": "HIGH",
  "likely_reason": "Status 'packed' was skipped in the order lifecycle — potential system bypass",
  "internal_next_action": "Verify the order processing pipeline. Missing status may indicate a system bypass or data integrity issue.",
  "confidence_score": 0.96
}

Example 2 — DUPLICATE_EVENT (LOW):
Input: Driver note says "hệ thống ghi trùng trạng thái". Event timeline shows: paid → paid (same status recorded twice).
Output:
{
  "exception_type": "DUPLICATE_EVENT",
  "severity": "LOW",
  "likely_reason": "Duplicate consecutive event detected: status 'paid' recorded multiple times",
  "internal_next_action": "Investigate the event source for duplicate submissions. No immediate action required.",
  "confidence_score": 0.98
}
