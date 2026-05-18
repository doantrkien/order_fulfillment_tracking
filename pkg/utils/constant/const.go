package constant

import "errors"

const (
	SUCCESS           = "Thành công.!"
	INVALID_INPUT     = "Dữ liệu đầu vào không hợp lệ.!"
	INVALID_STATUS    = "Trạng thái thay đổi không hợp lệ.!"
	ERROR             = "Có lỗi xảy ra trong quá trình xử lý.!"
	NOT_FOUND         = "Không tìm thấy dữ liệu.!"
	UN_AUTHENTICATION = "Vui lòng đăng nhập.!"
	UN_AUTHORIZATION  = "Bạn không có quyền thực hiện thao tác.!"
	URL_NOT_FOUND     = "Không tìm thấy địa chỉ phù hợp.!"
)

var (
	ERR_NOT_FOUND = errors.New("resource not found")
)
