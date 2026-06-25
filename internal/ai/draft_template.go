package ai

import "fmt"

var CustomerUpdateFallbackTemplates = map[string]string{
	// Group 1: Stuck Order
	"STUCK_ORDER": "Hello %s, your order is being processed slower than expected. We are actively coordinating with the shipping department to expedite the delivery. We sincerely apologize for this inconvenience.",

	// Group 2: Delivery Failure
	"DELIVERY_FAILURE": "Hello %s, we regret to inform you that an issue occurred during the shipment of your order to %s. The delivery team is currently investigating and will contact you as soon as possible.",

	// Group 3: Alternative Delivery
	"ALTERNATIVE_DELIVERY": "Hello %s, your order has been delivered and left at an alternative location. Please check with the reception or the person who received it on your behalf. If you have any concerns, please contact our support team.",

	// Group 4: Status Anomalies
	"INVALID_TRANSITION": "Hello %s, the system detected a status mismatch during your order update. Our technical team is verifying the information to ensure an accurate delivery route.",
	"SKIPPED_STATUS":     "Hello %s, we noticed an unusual update in your order processing progress. We are checking internally and will provide you with the most accurate update as soon as possible.",
	"DUPLICATE_EVENT":    "Hello %s, the system recorded a duplicate in your order status history. The operations department is resolving this inconsistency — your delivery progress will not be affected.",

	// Group 5: Default Fallback
	"OTHER":   "Hello %s, we are processing an issue that has arisen regarding your order. We will update you with detailed information and the next steps as soon as possible.",
	"DEFAULT": "Hello %s, we are verifying your order information due to an unexpected issue. Our support team will send you the latest update as soon as possible.",
}

func GetFallbackTemplate(exceptionType, customerName, shippingAddress string) string {
	template, exists := CustomerUpdateFallbackTemplates[exceptionType]
	if !exists {
		template = CustomerUpdateFallbackTemplates["DEFAULT"]
	}

	switch exceptionType {
	case "DELIVERY_FAILURE":
		return fmt.Sprintf(template, customerName, shippingAddress)
	default:
		return fmt.Sprintf(template, customerName)
	}
}
