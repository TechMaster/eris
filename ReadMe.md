# Cải tiến từ thư viện [https://github.com/rotisserie/eris](https://github.com/rotisserie/eris)

**Whole source code of this package is credited to rotisserie/eris.**

Ưu điểm lớn nhất của [rotisserie/eris](https://github.com/rotisserie/eris) đó là lỗi bao gồm cả stack trace giúp lập trình viên nhanh chóng tìm lỗi.
## Hướng dẫn sử dụng
### 1. Cài đặt package
```
go get -u github.com/TechMaster/eris
```

### 2. Tạo eris Error

#### 2.1 Tạo một lỗi cấp độ Error
```go
//Tạo một lỗi, thêm HTTP status code, trrar
func Bar() error {
	return eris.New("Không tìm thấy bản ghi trong CSDL").StatusCode(404).EnableJSON()
}
```



### 3. Xử lý lỗi eris Error
Kiểm tra lỗi trả về có kiểu là eris Error không
```go
var statusCode = 500
if e, ok := err.(*eris.Error); ok {
	handleEris(e)
	if e.Code > 0 { // Mặc định là 500, nếu e.Code > 0 thì gán vào statusCode
		statusCode = e.Code
	}
}
```

Hàm xử lý lỗi Eris
```go
//Hàm chuyên xử lý Eris Error có Stack Trace
func handleEris(err *eris.Error) {
	formattedStr := eris.ToCustomString(err, eris.StringFormat{
		Options: eris.FormatOptions{
			InvertOutput: true, // flag that inverts the error output (wrap errors shown first)
			WithTrace:    true, // flag that enables stack trace output
			InvertTrace:  true, // flag that inverts the stack trace output (top of call stack shown first)
			Skip:         3,    // Bỏ qua 3 dòng lệnh cuối cùng trong Stack
		},
		MsgStackSep:  "\n",  // separator between error messages and stack frame data
		PreStackSep:  "\t",  // separator at the beginning of each stack frame
		StackElemSep: " | ", // separator between elements of each stack frame
		ErrorSep:     "\n",  // separator between each error in the chain
	})

	colorReset := string("\033[0m")
	colorRed := string("\033[31m")
	//Chỗ này log ra console
	if err.IsPanic() {
		fmt.Println(colorRed, formattedStr, colorReset)
		//Lỗi Panic và Error nhất thiết phải ghi vào file !
	} else {
		fmt.Println(formattedStr)
	}
}
```

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