package eris

type ErrorType int //Cấp độ lỗi

const (
	WARNING  ErrorType = iota + 1 //Cảnh báo, ứng dụng vẫn chạy được
	ERROR                         //Lỗi, cần báo cho end user, log lỗi ra console
	SYSERROR                      //Lỗi hệ thống, báo cho end user Internal Server Error, log lỗi ra console và file
	PANIC                         //Lỗi nghiêm trọng, báo cho end user Internal Server Error, log lỗi ra console và file, thoát ứng dụng
)

type Error struct {
	global  bool                   // flag indicating whether the error was declared globally
	msg     string                 // root error message
	ext     error                  // error type for wrapping external errors
	stack   *stack                 // root error stack trace
	ErrType ErrorType              // Loại lỗi. Cường bổ xung
	Code    int                    // HTTP Status code. Cường bổ xung
	Data    map[string]interface{} //Thông tin bổ xung
}

func Warning(msg string) *Error {
	stack := callers(3) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     msg,
		stack:   stack,
		ErrType: WARNING,
	}
}

func New(msg string) *Error {
	stack := callers(3) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     msg,
		stack:   stack,
		ErrType: ERROR,
	}
}

func SysError(msg string) *Error {
	stack := callers(3) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     msg,
		stack:   stack,
		ErrType: SYSERROR,
	}
}

func Panic(msg string) *Error {
	stack := callers(3) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     msg,
		stack:   stack,
		ErrType: PANIC,
	}
}

func NewFrom(err error) *Error {
	stack := callers(3) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     err.Error(),
		stack:   stack,
		ErrType: ERROR,
	}
}

func WrapFrom(err error, skip int) *Error {
	stack := callers(skip) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     err.Error(),
		stack:   stack,
		ErrType: ERROR,
	}
}

//Bao lấy một error và thêm báo lỗi
func NewFromMsg(err error, msg string) *Error {
	stack := callers(3) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     err.Error() + " : " + msg,
		stack:   stack,
		ErrType: ERROR,
	}
}

func (error *Error) SetType(errType ErrorType) *Error {
	error.ErrType = errType
	return error
}

//Trả về mã lỗi HTTP error, thường áp dụng khi trả về request đến REST API
func (error *Error) StatusCode(statusCode int) *Error {
	error.Code = statusCode
	return error
}

//Gán mã lỗi khi request gửi lên không hợp lệ
func (error *Error) BadRequest() *Error {
	error.Code = 400
	return error
}

//Gán mã lỗi khi người dùng không được phép truy cập
func (error *Error) UnAuthorized() *Error {
	error.Code = 401
	return error
}

//Không tìm thấy một record theo yêu cầu có thể trả về lỗi này
func (error *Error) NotFound() *Error {
	error.Code = 404
	return error
}

//Dành cho hầu hết lỗi phát sinh phía server
func (error *Error) InternalServerError() *Error {
	error.Code = 500
	return error
}

//Thêm thông tin bổ xung vào lỗi để client có thể xử lý thêm
func (error *Error) SetData(data map[string]interface{}) *Error {
	error.Data = data
	return error
}

//Truyền vào một error bất kỳ kiểm tra xem có phải là lỗi hệ thống
func IsSysError(err error) bool {
	if e, ok := err.(*Error); ok && e.ErrType == SYSERROR {
		return true
	} else {
		return false
	}
}

//Truyền vào một error bất kỳ kiểm tra xem có phải là lỗi nghiêm trọng
func IsPanic(err error) bool {
	if e, ok := err.(*Error); ok && e.ErrType == PANIC {
		return true
	} else {
		return false
	}
}
