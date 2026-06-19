package ai

// CustomerUpdateFallbackTemplates maps each internal exception type to a static,
// customer-friendly default message template.
// These templates use placeholders [REDACTED_CUSTOMER_NAME] and [REDACTED_SHIPPING_ADDRESS]
// for safe, server-side data masking and hydration at the client.
var CustomerUpdateFallbackTemplates = map[string]string{
	// Group 1: Stuck Order (STUCK_ORDER)
	"STUCK_ORDER": "Hello [REDACTED_CUSTOMER_NAME], your order is being processed slower than expected. We are actively coordinating with the shipping department to expedite the delivery. We sincerely apologize for this inconvenience.",
	
	// Group 2: Delivery Incident (DELIVERY_FAILURE)
	"DELIVERY_FAILURE": "Hello [REDACTED_CUSTOMER_NAME], we regret to inform you that an issue occurred during the shipment of your order to [REDACTED_SHIPPING_ADDRESS]. The delivery team is currently investigating and will contact you as soon as possible.",
	
	// Group 3: Status Anomalies (INVALID_TRANSITION, SKIPPED_STATUS, DUPLICATE_EVENT, CANCELLATION_ANOMALY, REFUND_ANOMALY)
	"INVALID_TRANSITION":   "Hello [REDACTED_CUSTOMER_NAME], the system detected a status mismatch during your order update. Our technical team is verifying the information to ensure an accurate delivery route.",
	"SKIPPED_STATUS":       "Hello [REDACTED_CUSTOMER_NAME], we noticed an unusual update in your order processing progress. We are checking internally and will provide you with the most accurate update as soon as possible.",
	"DUPLICATE_EVENT":      "Hello [REDACTED_CUSTOMER_NAME], the system recorded a duplicate in your order status history. The operations department is resolving this inconsistency — your delivery progress will not be affected.",
	"CANCELLATION_ANOMALY": "Hello [REDACTED_CUSTOMER_NAME], we detected an unusual cancellation status or an invalid cancellation request regarding your order. Our Customer Support team is verifying it and will contact you directly as soon as possible.",
	"REFUND_ANOMALY":       "Hello [REDACTED_CUSTOMER_NAME], we detected an anomaly related to your refund request. The finance department is processing the details to ensure your maximum protection.",

	// Group 4: Default Fallback (OTHER or unclassified errors)
	"OTHER":   "Hello [REDACTED_CUSTOMER_NAME], we are processing an issue that has arisen regarding your order. We will update you with detailed information and the next steps as soon as possible.",
	"DEFAULT": "Hello [REDACTED_CUSTOMER_NAME], we are verifying your order information due to an unexpected issue. Our support team will send you the latest update as soon as possible.",
}

// GetFallbackTemplate returns the static message template for a given exception type.
// If the exception type does not exist in the map, it returns the DEFAULT template.
func GetFallbackTemplate(exceptionType string) string {
	if template, exists := CustomerUpdateFallbackTemplates[exceptionType]; exists {
		return template
	}
	return CustomerUpdateFallbackTemplates["DEFAULT"]
}
