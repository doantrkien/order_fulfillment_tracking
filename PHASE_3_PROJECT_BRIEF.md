## 1. TỔNG QUAN & MỤC TIÊU (EXECUTIVE SUMMARY)

Hệ thống Phase 2 của team hiện đang quản lý các bước chuyển đổi trạng thái đơn hàng, xác thực sự kiện, xử lý cập nhật từ tài xế và xuất báo cáo vận hành mỗi ngày.

**Mục tiêu nâng cấp Phase 3:** Biến hệ thống hiện tại thành một "Trợ lý ngoại lệ" (Exception-assist system) có khả năng tự động giải thích các đơn hàng bất thường và soạn thảo các thông báo cập nhật an toàn cho khách hàng.

**4 nhiệm vụ cốt lõi của AI:**
1. Phát hiện các ngoại lệ trong quá trình hoàn tất đơn hàng.
2. Giải thích nguyên nhân khả thi dẫn đến ngoại lệ.
3. Đề xuất hành động xử lý cho bộ phận vận hành.
4. Soạn thảo tin nhắn cập nhật cho khách hàng (tuyệt đối không tự động gửi).

**Nguyên tắc phân phối:** Phase 3 được đánh giá như một Dự án phần mềm thực tế, không phải một bản demo AI dùng thử. Team tuyệt đối không nhét các tính năng AI ngẫu nhiên vào hệ thống. AI chỉ được thêm vào đúng nơi nó giúp cải thiện luồng công việc, có thể giải thích được logic với reviewer và phải có bằng chứng kiểm thử rõ ràng.

## 2. PHÂN QUYỀN NGUỒN LỰC (USER PERSONAS)

- **Fulfillment Operator (Nhân viên vận hành):** Người trực tiếp sử dụng hệ thống để nhận diện các đơn hàng đang gặp sự cố cần được can thiệp.
- **Customer Support Agent (CSKH):** Người đọc, kiểm duyệt và quyết định sử dụng bản nháp tin nhắn do AI soạn thảo để gửi cho khách.
- **Operations Manager (Quản lý vận hành):** Người theo dõi các chuỗi/mẫu ngoại lệ (exception patterns) và đánh giá hiệu suất giao hàng tổng thể.

## 3. HỢP ĐỒNG DỮ LIỆU (INPUTS & OUTPUTS)

AI Capability của Team 2 giao tiếp qua API với JSON Contract nghiêm ngặt.

### Input Context (Dữ liệu nạp vào AI)
- Chi tiết đơn hàng, trạng thái hiện tại, các mốc thời gian (timestamps) và tình trạng thanh toán/hoàn tiền.
- Dòng thời gian sự kiện (event timeline) và ghi chú giao hàng từ các tài xế giả lập (synthetic drivers).
- Dữ liệu tổng hợp báo cáo đơn hàng hàng ngày và ngưỡng mốc thời gian giao hàng dự kiến.
- Tùy chọn: Các tài liệu hướng dẫn văn phong tin nhắn CSKH (synthetic text).

## 4. YÊU CẦU KỸ THUẬT BACKEND (TECHNICAL BASELINE)

### 4.1. Phạm vi kế thừa từ Phase 2
- Tái sử dụng repo scaffolding, pattern xác thực/token, logging/Request ID, error model và Docker Compose.
- Tái sử dụng kiến trúc Go với package boundaries chuẩn: `cmd`, `internal/api`, `internal/domain`, `internal/repository`, `internal/worker`, `internal/config`, và `internal/observability`.
- Giữ nguyên cơ chế Worker-pool xử lý concurrency từ Phase 2. Các job AI chạy ngầm không được tạo ra race condition hoặc gọi model song song không kiểm soát.

### 4.2. Mở rộng Database (PostgreSQL)
Phải giữ nguyên PostgreSQL làm Source of Truth với các transaction boundaries chuẩn. Thiết kế migrations tạo thêm 3 bảng mới:
- `ai_order_exception_results`: Lưu lịch sử phân tích ngoại lệ, phiên bản prompt, điểm confidence và timestamps.
- `ai_customer_update_drafts`: Lưu các bản nháp CSKH được tạo.
- `ai_evaluation_runs`: Lưu trữ lịch sử các lần chạy bộ test đánh giá model.

### 4.3. Danh sách REST API Endpoints cần xây dựng
Tính năng AI bắt buộc phải bọc sau REST API, không được gọi model trực tiếp bằng file script.
- `POST /ai/orders/{id}/exception-analysis`: Kích hoạt phân tích ngoại lệ cho một đơn hàng cụ thể.
- `GET /ai/orders/{id}/insights/latest`: Truy xuất kết quả phân tích AI mới nhất của đơn hàng.
- `POST /ai/orders/customer-update-draft`: Yêu cầu AI tạo bản nháp CSKH độc lập.
- `POST /ai/evaluations/order-exceptions`: Trigger chạy tự động bộ dữ liệu đánh giá AI (Evaluation suite).

### 4.4. Pattern Kiến trúc AI đề xuất
- **Luồng xử lý:** `Controller/API` -> `Input validation` -> `Domain context builder` -> `AI adapter` -> `Output schema validator` -> `Confidence/fallback policy` -> `Audit log` -> `API response`.
- **Giao tiếp Mock/Stub:** Bắt buộc bọc model sau một AI Adapter Interface nội bộ. Khi chạy Unit Test, hệ thống phải trỏ được về một "Fake model implementation" trả ra kết quả deterministic.
- **Configuration-driven:** Tên Provider, Model, API Key, Timeout, Retry limit, Max input size và cờ Feature_Toggle (bật/tắt AI) bắt buộc nạp từ Environment Variables, cấm hardcode.

## 5. RANH GIỚI AN TOÀN & NON-GOALS (GUARDRAILS)

### 5.1. Strict Non-Goals (Không làm)
- **Không làm UI Frontend:** Chỉ cần trả API chuẩn, demo bằng Postman/cURL hoặc file Docs tự gen là đạt.
- **Không Over-engineering:** Không triển khai Kubernetes, Kafka/RabbitMQ, Data Warehouse hay Enterprise RBAC.
- **Không dùng Real Data:** Tuyệt đối không đưa thông tin định danh (PII) thật của khách hàng vào Prompt. Bắt buộc phải che (masking) định danh trước khi đẩy context cho AI.

### 5.2. Ranh giới Đỏ nghiệp vụ (Business Guardrails)
- **Cấm tự động hành động:** AI được phép soạn nháp tin nhắn, nhưng tuyệt đối không được tự động bấm gửi tin nhắn cho khách, không được tự động hoàn tiền (refund) và không được tự đổi trạng thái đơn hàng.
- **Ưu tiên Deterministic Logic:** Việc chặn các bước chuyển trạng thái đơn hàng sai quy tắc (invalid state transitions) là trách nhiệm của code logic Go thông thường, không do AI phán xét.
- **Kiểm soát nội dung CSKH:** Bản nháp tin nhắn do AI tạo ra tuyệt đối không được chứa các lời hứa hẹn đền bù vô căn cứ và cấm để lộ các chi tiết lỗi kỹ thuật nội bộ (ví dụ: cấm nói “do server PostgreSQL bị sập”).

### 5.3. Cơ chế dự phòng (Non-AI Baseline)
Khi tắt cờ AI (Feature disabled) hoặc AI bị timeout/trả về JSON lỗi, hệ thống phải tự động rơi về cơ chế Heuristic thuần code. Cơ chế này dùng các quy tắc code sẵn dựa vào: tuổi đơn hàng, các mốc milestone bị thiếu, nỗ lực chuyển trạng thái sai và sự kiện trùng lặp để tự điền các trường `exception_type` và `severity`.
