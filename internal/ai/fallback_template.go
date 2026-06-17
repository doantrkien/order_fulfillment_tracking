package ai

// CustomerUpdateFallbackTemplates maps each internal exception type to a static,
// customer-friendly default message template.
// These templates use placeholders [REDACTED_CUSTOMER_NAME] and [REDACTED_SHIPPING_ADDRESS]
// for safe, server-side data masking and hydration at the client.
var CustomerUpdateFallbackTemplates = map[string]string{
	// Group 1: Stuck Orders / Delayed Delivery (STUCK_ORDER)
	"STUCK_ORDER": "Dear [REDACTED_CUSTOMER_NAME], your order is currently processing slower than expected. We are actively working with our logistics department to expedite the delivery. We sincerely apologize for this inconvenience.",

	// Group 2: Transit Issues / Delivery Failure (DELIVERY_FAILURE)
	"DELIVERY_FAILURE": "Dear [REDACTED_CUSTOMER_NAME], we regret to inform you that an unexpected issue has occurred during transit to the address [REDACTED_SHIPPING_ADDRESS]. Our dispatchers are investigating and will contact you as soon as possible.",
	
	// Group 3: Status Transitions / System Event Anomalies
	"INVALID_TRANSITION":   "Dear [REDACTED_CUSTOMER_NAME], our system has detected a mismatch in your order status update. Our technical team is reviewing the information to ensure an accurate delivery route.",
	"SKIPPED_STATUS":       "Dear [REDACTED_CUSTOMER_NAME], we have noticed some unusual updates in your order status. We are conducting an internal audit to provide you with the most accurate information shortly.",
	"DUPLICATE_EVENT":      "Dear [REDACTED_CUSTOMER_NAME], a duplication has been recorded in your order's history. Our operations team is resolving the data discrepancy, and your delivery timeline will not be affected.",
	"CANCELLATION_ANOMALY": "Dear [REDACTED_CUSTOMER_NAME], we have detected an unusual cancellation request or status regarding your order. Our Customer Support team is verifying this and will contact you directly at the earliest.",
	"REFUND_ANOMALY":       "Dear [REDACTED_CUSTOMER_NAME], we have identified a data anomaly related to your refund request. Our finance department is currently processing it to ensure your full benefits are protected.",

	// Group 4: Default Fallback Templates (OTHER or Unclassified Errors)
	"OTHER":   "Dear [REDACTED_CUSTOMER_NAME], we are thoroughly investigating an unexpected issue related to your order. We will provide detailed updates and the next steps as soon as possible.",
	"DEFAULT": "Dear [REDACTED_CUSTOMER_NAME], we are currently verifying your order details due to an unforeseen issue. Our support team will send you the latest updates within the shortest time possible.",
}

// GetFallbackTemplate returns the static message template for a given exception type.
// If the exception type does not exist in the map, it returns the DEFAULT template.
func GetFallbackTemplate(exceptionType string) string {
	if template, exists := CustomerUpdateFallbackTemplates[exceptionType]; exists {
		return template
	}
	return CustomerUpdateFallbackTemplates["DEFAULT"]
}
