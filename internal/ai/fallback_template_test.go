package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFallbackTemplate(t *testing.T) {
	tests := []struct {
		exceptionType string
		wantContains  string
	}{
		{"STUCK_ORDER", "chậm hơn dự kiến"},
		{"DELIVERY_FAILURE", "địa chỉ [REDACTED_SHIPPING_ADDRESS]"},
		{"INVALID_TRANSITION", "trạng thái không khớp"},
		{"SKIPPED_STATUS", "cập nhật bất thường"},
		{"DUPLICATE_EVENT", "sự trùng lặp"},
		{"CANCELLATION_ANOMALY", "trạng thái hủy đơn bất thường"},
		{"REFUND_ANOMALY", "yêu cầu hoàn tiền"},
		{"OTHER", "sự cố phát sinh liên quan đến đơn hàng"},
		{"SOME_RANDOM_TYPE", "phát sinh ngoài ý muốn"}, // Dẫn về DEFAULT
		{"", "phát sinh ngoài ý muốn"},                 // Dẫn về DEFAULT
	}

	for _, tc := range tests {
		t.Run(tc.exceptionType, func(t *testing.T) {
			got := GetFallbackTemplate(tc.exceptionType)
			assert.Contains(t, got, "[REDACTED_CUSTOMER_NAME]")
			assert.Contains(t, got, tc.wantContains)
		})
	}
}
