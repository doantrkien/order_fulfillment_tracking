package constant

import "errors"

const (
	SUCCESS       = "Thành công.!"
	INVALID_INPUT = "Dữ liệu đầu vào không hợp lệ.!"
	ERROR         = "Có lỗi xảy ra trong quá trình xử lý.!"
	NOT_FOUND     = "Không tìm thấy dữ liệu.!"
)

var (
	ERR_NOT_FOUND = errors.New("resource not found")
)
