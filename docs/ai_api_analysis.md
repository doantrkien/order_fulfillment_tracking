# AI Exception Analysis API — Tổng hợp Nghiệp vụ & Kiến trúc

## 1. Tổng quan
**Endpoint:** `POST /api/v1/ai/orders/{id}/exception-analysis`  
**Người dùng:** Quản lý kho / Nhân viên CSKH (gọi thủ công từ Dashboard)  
**Mục tiêu:** Phân tích đơn hàng để phát hiện lỗi, phán đoán nguyên nhân và gợi ý hướng xử lý.

---

## 2. Flow xử lý

```
Request (Manager bấm nút)
    │
    ▼
[Handler] Parse orderID, Parse body.note
    │
    ▼
[Service] GetAIContextByOrderID → Load toàn bộ đơn hàng + Events từ DB
    │
    ▼
[ExceptionAnalyzer.Analyze]
    │
    ├─ Step 1: Chạy Rule-based Engine → chuẩn bị sẵn ruleResult (phòng hờ)
    │
    ├─ Step 2: hasDriverNote? (Events.DriverNote != "" || notes != "")
    │              │
    │              ▼ KHÔNG có note → Return ruleResult (ai_disabled / no_driver_note)
    │
    ├─ Step 3: AI Enabled?
    │              │
    │              ▼ KHÔNG → Return ruleResult (ai_disabled)
    │
    └─ Step 4: Gọi AI Adapter
                   │
                   ▼
           [AIAdapter.AnalyzeException]
                   │
                   ├─ ClassifyDriverNote(notes) → Lấy Knowledge Base theo từ khóa
                   │   ├─ Matched → KB nhỏ (1-2 cuốn, ~9000 chars)
                   │   └─ Không match → KB đầy đủ (5 cuốn, ~18000 chars)
                   │
                   ├─ BuildExceptionAnalysisPrompt(ctx, knowledge)
                   │
                   └─ GenerateContent → Groq/Gemini LLM
                              │
                              ├─ Lỗi mạng/Timeout → Fallback → ruleResult
                              ├─ JSON sai cấu trúc → Fallback → ruleResult
                              ├─ ConfidenceScore < 0.6 → Fallback → ruleResult
                              └─ ✅ Thành công → Trả về kết quả AI
```

---

## 3. Request / Response

### Request
```json
POST /api/v1/ai/orders/{id}/exception-analysis
Body (tùy chọn):
{
  "note": "Ghi chú thêm của quản lý (tối đa 500 ký tự)"
}
```

### Response
```json
{
  "status": "SUCCESS",
  "data": {
    "result_id": "res-42",
    "order_id": "1",
    "exception_type": "DELIVERY_FAILURE",
    "severity": "HIGH",
    "likely_reason": "Vehicle breakdown prevented delivery",
    "internal_next_action": "Escalate to logistics team. Arrange re-delivery.",
    "confidence_score": 0.92,
    "fallback_used": false,
    "prompt_template_version": "v1.0.0",
    "evaluated_at": "2026-06-24T08:30:00Z"
  }
}
```

---

## 4. Rule-based Engine — Các loại lỗi được nhận diện

Thứ tự ưu tiên (chạy từ trên xuống, dừng lại khi tìm thấy lỗi đầu tiên):

| Hàm | Điều kiện kích hoạt | ExceptionType |
|---|---|---|
| `detectDeliveryFailure` | Status = SHIPPED + có từ khóa trong Driver Note | DELIVERY_FAILURE |
| `detectDuplicateEvents` | Lịch sử Events có trạng thái bị ghi 2 lần | DUPLICATE_EVENT |
| `detectSkippedStatuses` | Đơn nhảy cóc qua 1+ trạng thái bắt buộc | SKIPPED_STATUS |
| `detectInvalidTransitions` | Đơn chuyển trạng thái không hợp lệ | INVALID_TRANSITION |
| `detectStuckOrder` | Đơn bị kẹt lâu hơn ngưỡng cho phép theo trạng thái | STUCK_ORDER |
| *(không tìm thấy gì)* | — | nil → `OTHER / LOW` |

### Ngưỡng thời gian bị kẹt (Stuck Order Threshold)
| Trạng thái | Ngưỡng | Mức độ (theo ratio) |
|---|---|---|
| CREATED | 24 giờ | ratio ≤1.25 → LOW, ≤2.0 → MEDIUM, ≤3.0 → HIGH, >3.0 → CRITICAL |
| PAID | 48 giờ | (tương tự) |
| PACKED | 24 giờ | (tương tự) |
| SHIPPED | 72 giờ | (tương tự) |

---

## 5. Knowledge Base — Phân loại theo từ khóa (ClassifyDriverNote)

Khi không có từ khóa nào khớp → gửi **toàn bộ 5 cuốn** → có thể gây Prompt Too Large nếu `AI_MAX_INPUT_SIZE` < 18000.

| Nhóm KB | Từ khóa mẫu |
|---|---|
| Delivery Failure | lost, stolen, xe hỏng, tai nạn, sai địa chỉ, không có nhà, khách không nghe máy |
| Duplicate Event | duplicate, lặp, trùng, repeated |
| Skipped Status | skipped, missed, bỏ qua, thiếu |
| Stuck Order | stuck, delay, late, trễ, chậm |
| Order State Machine | *(Luôn được gắn vào — mọi lần)* |

---

## 6. Chiến lược Fallback (Phòng thủ nhiều lớp)

| Lý do Fallback | Mã | Kết quả trả về |
|---|---|---|
| AI bị tắt cấu hình | `ai_disabled` | ruleResult hoặc OTHER/LOW |
| Không có ghi chú tài xế | `no_driver_note` | ruleResult |
| Prompt quá lớn (vượt MaxInputSize) | `ai_connection_error` (classify) | ruleResult |
| Lỗi kết nối/Timeout | `ai_connection_error` / `ai_timeout` | ruleResult |
| AI trả về JSON sai cấu trúc | `ai_invalid_response` | ruleResult |
| Confidence Score < 0.6 | `ai_confidence_below_threshold` | ruleResult |

**Khi `ruleResult == nil`** (không phát hiện lỗi nào), fallback cuối cùng là:
```
ExceptionType: "OTHER", Severity: "LOW"
InternalNextAction: "No action required. Monitor the order for further changes."
```

---

## 7. Cài đặt môi trường quan trọng

| Biến môi trường | Giá trị mặc định | Ý nghĩa |
|---|---|---|
| `AI_ENABLED` | true | Bật/tắt tính năng AI |
| `AI_MAX_INPUT_SIZE` | 4000 (code) / 10000 (env) | Giới hạn ký tự prompt. Cần ≥ 18000 để cover mọi note |
| `AI_TIMEOUT_MS` | 10000ms (10s) | Timeout mỗi lần gọi AI |
| `AI_RETRY_LIMIT` | 3 | Số lần retry khi AI thất bại |
| `GROQ_MODEL` | llama-3.3-70b-versatile | Model LLM đang dùng |

---

## 8. Output được lưu vào Database (bảng `ai_exceptions`)

Mỗi lần API được gọi, kết quả sẽ được lưu lại (kể cả khi Fallback). Các field lưu:
- `exception_type`, `severity`, `likely_reason`, `internal_next_action`
- `confidence_score`, `fallback_used`, `fallback_reason`
- `prompt_template_version`, `duration_ms`, `raw_response`
- `evaluated_at`

---

---

# API Soạn Tin Nhắn Khách Hàng (Generate Draft)

## 1. Tổng quan
**Endpoint:** `POST /api/v1/ai/orders/customer-update-draft`  
**Người dùng:** Quản lý kho / CSKH  
**Mục tiêu:** Dựa vào kết quả phân tích lỗi đơn hàng (đã lưu trong DB), tự động soạn thảo một tin nhắn thông báo cho khách hàng.

**Điều kiện tiên quyết:** Phải đã gọi API `exception-analysis` trước để có kết quả phân tích trong DB.

---

## 2. Flow xử lý

```
Request (Manager chọn Tone & Channel, bấm "Soạn thảo")
    │
    ▼
[Handler] Parse body: {order_id, tone, channel}
    │
    ▼
[Service.GenerateDraft]
    ├─ GetAIContextByOrderID → Load thông tin đơn hàng
    ├─ GetLatestAnalysisByOrderID → Load kết quả phân tích lỗi mới nhất
    │
    ▼
[DraftGenerator.Generate]
    │
    ├─ buildFallbackDraftMessage → Chuẩn bị sẵn Template cứng theo ExceptionType
    │
    ├─ shouldCallAI?
    │   ├─ ExceptionType == "OTHER" → CẦN AI (cần diễn giải bằng ngôn ngữ tự nhiên)
    │   ├─ Tone không phải neutral/informative → CẦN AI (apologetic, proactive...)
    │   ├─ Channel == "sms" → CẦN AI (cần rút gọn)
    │   └─ Còn lại → KHÔNG CẦN AI → Dùng Template cứng luôn (tiết kiệm tiền)
    │
    └─ (Nếu cần AI) Gọi AIAdapter.DraftCustomerUpdate
                │
                ├─ Lỗi mạng/Timeout → Fallback → Template cứng
                ├─ JSON sai/rỗng → Fallback → Template cứng
                ├─ ConfidenceScore < 0.6 → Fallback → Template cứng
                └─ ✅ Thành công → Trả về tin nhắn do AI soạn
```

---

## 3. Request / Response

### Request
```json
POST /api/v1/ai/orders/customer-update-draft
{
  "order_id": 123,
  "tone": "apologetic",
  "channel": "email"
}
```

**Các giá trị hợp lệ:**
- `tone`: `neutral` | `informative` | `apologetic` | `proactive` *(hoặc "empathetic" → tự động map sang "apologetic")*
- `channel`: `email` | `sms` | *(bỏ trống)*

### Response
```json
{
  "status": "SUCCESS",
  "data": {
    "order_id": 123,
    "draft_message": "Kính gửi Anh/Chị Nguyễn Văn A...",
    "tone": "apologetic",
    "confidence_score": 0.95,
    "fallback_used": false,
    "prompt_template_version": "v1.0.0",
    "generated_at": "2026-06-24T09:00:00Z"
  }
}
```

---

## 4. Template Fallback theo loại lỗi

Khi không cần AI (hoặc AI thất bại), hệ thống dùng template cứng theo `exception_type`:

| ExceptionType | Template mẫu |
|---|---|
| STUCK_ORDER | "...your order is being processed slower than expected..." |
| DELIVERY_FAILURE | "...an issue occurred during the shipment of your order..." |
| INVALID_TRANSITION | "...the system detected a status mismatch..." |
| SKIPPED_STATUS | "...we noticed an unusual update in your order processing..." |
| DUPLICATE_EVENT | "...the system recorded a duplicate in your order status history..." |
| OTHER / DEFAULT | "...we are processing an issue that has arisen regarding your order..." |

---

## 5. Kết quả lưu vào DB (bảng `ai_customer_update_drafts`)

- `order_id`, `ai_exception_result_id` (FK sang bảng phân tích)
- `draft_message`, `tone`, `channel`
- `confidence_score`, `fallback_used`, `fallback_reason`
- `review_status` (mặc định: `pending` — chờ người dùng review trước khi gửi)
- `raw_response`, `prompt_template_version`, `duration_ms`
