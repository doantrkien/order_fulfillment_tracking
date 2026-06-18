package ai

// CustomerUpdateFallbackTemplates maps each internal exception type to a static,
// customer-friendly default message template.
// These templates use placeholders [REDACTED_CUSTOMER_NAME] and [REDACTED_SHIPPING_ADDRESS]
// for safe, server-side data masking and hydration at the client.
var CustomerUpdateFallbackTemplates = map[string]string{
	// Group 1: Đơn hàng bị trễ / Stuck Order (STUCK_ORDER)
	"STUCK_ORDER": "Xin chào [REDACTED_CUSTOMER_NAME], đơn hàng của bạn đang xử lý chậm hơn dự kiến. Chúng tôi đang tích cực phối hợp với bộ phận vận chuyển để đẩy nhanh tiến độ. Rất xin lỗi vì sự bất tiện này.",
	// Group 2: Sự cố giao hàng (DELIVERY_FAILURE)
	"DELIVERY_FAILURE": "Xin chào [REDACTED_CUSTOMER_NAME], chúng tôi rất tiếc phải thông báo rằng đã xảy ra sự cố trong quá trình vận chuyển đơn hàng đến địa chỉ [REDACTED_SHIPPING_ADDRESS]. Đội ngũ giao hàng đang điều tra và sẽ liên hệ với bạn sớm nhất có thể.",
	// Group 3: Sự kiện bất thường về trạng thái (INVALID_TRANSITION, SKIPPED_STATUS, DUPLICATE_EVENT, CANCELLATION_ANOMALY, REFUND_ANOMALY)
	"INVALID_TRANSITION":   "Xin chào [REDACTED_CUSTOMER_NAME], hệ thống phát hiện trạng thái không khớp trong quá trình cập nhật đơn hàng của bạn. Đội kỹ thuật đang xác minh thông tin để đảm bảo lộ trình giao hàng chính xác.",
	"SKIPPED_STATUS":       "Xin chào [REDACTED_CUSTOMER_NAME], chúng tôi nhận thấy có cập nhật bất thường trong tiến trình xử lý đơn hàng của bạn. Chúng tôi đang kiểm tra nội bộ và sẽ cập nhật thông tin chính xác nhất đến bạn trong thời gian sớm nhất.",
	"DUPLICATE_EVENT":      "Xin chào [REDACTED_CUSTOMER_NAME], hệ thống ghi nhận sự trùng lặp trong lịch sử trạng thái đơn hàng. Bộ phận vận hành đang xử lý sự không nhất quán này — tiến độ giao hàng của bạn sẽ không bị ảnh hưởng.",
	"CANCELLATION_ANOMALY": "Xin chào [REDACTED_CUSTOMER_NAME], chúng tôi phát hiện trạng thái hủy đơn bất thường hoặc yêu cầu hủy không hợp lệ liên quan đến đơn hàng của bạn. Bộ phận Chăm sóc Khách hàng đang xác minh và sẽ liên hệ trực tiếp với bạn sớm nhất có thể.",
	"REFUND_ANOMALY":       "Xin chào [REDACTED_CUSTOMER_NAME], chúng tôi phát hiện bất thường liên quan đến yêu cầu hoàn tiền của bạn. Bộ phận tài chính đang xử lý thông tin để đảm bảo quyền lợi tối đa cho bạn.",

	// Group 4: Fallback mặc định (OTHER hoặc lỗi chưa phân loại)
	"OTHER":   "Xin chào [REDACTED_CUSTOMER_NAME], chúng tôi đang xử lý sự cố phát sinh liên quan đến đơn hàng của bạn. Chúng tôi sẽ cập nhật thông tin chi tiết và các bước tiếp theo sớm nhất có thể.",
	"DEFAULT": "Xin chào [REDACTED_CUSTOMER_NAME], chúng tôi đang xác minh thông tin đơn hàng của bạn do có sự cố phát sinh ngoài ý muốn. Đội hỗ trợ sẽ gửi cập nhật mới nhất đến bạn trong thời gian sớm nhất.",
}

// GetFallbackTemplate returns the static message template for a given exception type.
// If the exception type does not exist in the map, it returns the DEFAULT template.
func GetFallbackTemplate(exceptionType string) string {
	if template, exists := CustomerUpdateFallbackTemplates[exceptionType]; exists {
		return template
	}
	return CustomerUpdateFallbackTemplates["DEFAULT"]
}
