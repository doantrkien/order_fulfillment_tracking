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