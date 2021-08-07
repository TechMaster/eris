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
	ErrType ErrorType              //Loại lỗi. Cường bổ xung
	Code    int                    // HTTP Status code. Cường bổ xung
	JSON    bool                   // true if error will resonse as JSON for REST request, false to render error page at server side
	Data    map[string]interface{} //Thông tin bổ xung
}

func Warning(msg string) *Error {
	stack := callers(3) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     msg,
		stack:   stack,
		ErrType: WARNING,
		JSON:    false, //server side rendered error page not return JSON error
	}
}

func New(msg string) *Error {
	stack := callers(3) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     msg,
		stack:   stack,
		ErrType: ERROR,
		JSON:    false,
	}
}

func SysError(msg string) *Error {
	stack := callers(3) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     msg,
		stack:   stack,
		ErrType: SYSERROR,
		JSON:    false,
	}
}

func Panic(msg string) *Error {
	stack := callers(3) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     msg,
		stack:   stack,
		ErrType: PANIC,
		JSON:    false,
	}
}

func NewFrom(err error) *Error {
	stack := callers(3) // callers(3) skips this method, stack.callers, and runtime.Callers
	return &Error{
		global:  stack.isGlobal(),
		msg:     err.Error(),
		stack:   stack,
		ErrType: ERROR,
		JSON:    false,
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
		JSON:    false,
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

//Lỗi trả về dạng JSON reponse bao gồm status code mặc định 500
func (error *Error) EnableJSON() *Error {
	error.JSON = true
	if error.Code == 0 {
		error.Code = 500 //Internal server error by default
	}
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
