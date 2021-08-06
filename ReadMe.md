# Cải tiến từ thư viện [https://github.com/rotisserie/eris](https://github.com/rotisserie/eris)

**Whole source code of this package is credited to rotisserie/eris.**

## Cường đã tạo ra những thay đổi sau đây

### 1. Đổi `rootError` thành `Error`
```go
type rootError struct {
	global bool   // flag indicating whether the error was declared globally
	msg    string // root error message
	ext    error  // error type for wrapping external errors
	stack  *stack // root error stack trace
}
```

```go
type Error struct {
	global  bool      // flag indicating whether the error was declared globally
	msg     string    // root error message
	ext     error     // error type for wrapping external errors
	stack   *stack    // root error stack trace
	ErrType ErrorType //Loại lỗi. Cường bổ xung
	Code    int       // HTTP Status code. Cường bổ xung
}
```

### 2. Thêm code vào eris.go

```go
type ErrorType int

const (
	WARNING  ErrorType = iota + 1 //Cảnh báo, ứng dụng vẫn chạy được
	ERROR                         //Lỗi, cần báo cho end user, log lỗi ra console
	SYSERROR                      //Lỗi hệ thống, báo cho end user Internal Server Error, log lỗi ra console và file
	PANIC                         //Lỗi nghiêm trọng, báo cho end user Internal Server Error, log lỗi ra console và file, thoát ứng dụng
)

type Error struct {
	global  bool      // flag indicating whether the error was declared globally
	msg     string    // root error message
	ext     error     // error type for wrapping external errors
	stack   *stack    // root error stack trace
	ErrType ErrorType //Loại lỗi. Cường bổ xung
	Code    int       // HTTP Status code. Cường bổ xung
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
func (error *Error) StatusCode(statusCode int) *Error {
	error.Code = statusCode
	return error
}
func (error *Error) IsSysError() bool {
	return error.ErrType == SYSERROR
}

func (error *Error) IsPanic() bool {
	return error.ErrType == PANIC
}
```

### 3. Thêm trường Skip vào FormatOptions ở format.go

```go
type FormatOptions struct {
	InvertOutput bool // Flag that inverts the error output (wrap errors shown first).
	WithTrace    bool // Flag that enables stack trace output.
	InvertTrace  bool // Flag that inverts the stack trace output (top of call stack shown first).
	WithExternal bool // Flag that enables external error output.
	Skip         int  // Cuong: Bỏ bớt một số hàm đầu tiên
}
```

Vừa sửa hàm này để bỏ qua Skip phương thức trong Stack Trace chủ yế
```go
func (err *ErrRoot) formatStr(format StringFormat) string {
	str := err.Msg + format.MsgStackSep
	if format.Options.WithTrace {
		stackArr := err.Stack.format(format.StackElemSep, format.Options.InvertTrace)
    // Cường thêm để bỏ qua Skip phần tử cuối cùng trong Stack Trace
		if len(stackArr) > format.Options.Skip+1 {
			stackArr = stackArr[:len(stackArr)-format.Options.Skip]
		}

		for i, frame := range stackArr {
			str += format.PreStackSep + frame
			if i < len(stackArr)-1 {
				str += format.ErrorSep
			}
		}
	}
	return str
}
```