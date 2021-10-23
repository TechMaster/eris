# Cải tiến từ thư viện [https://github.com/rotisserie/eris](https://github.com/rotisserie/eris)

## 1. Ưu điểm bắt lỗi, xử lý lỗi với [TechMaster/eris](https://github.com/TechMaster/eris)

Ưu điểm lớn nhất của [rotisserie/eris](https://github.com/rotisserie/eris) đó là lỗi bao gồm cả stack trace giúp lập trình viên nhanh chóng tìm lỗi. Tuy nhiên rotisserie/eris còn hạn chế:
1. Thiếu thuộc tính báo cấp độ lỗi
2. Cấu trúc Error để private nên khó truy cập thuộc tính bên trong
3. Chứa có cú pháp Fluent API để nối chuỗi các hàm tạo lỗi.

TechMaster/eris bổ xung những chức năng còn thiếu trên.

## 2. Cài đặt

Trong terminal ở thư mục dự án Golang hãy gõ
```
go get -u github.com/TechMaster/eris
```

Trong ứng dụng, import bằng lệnh
```go
import(
	"github.com/TechMaster/eris"
)
```

## 3. Chi tiết phần cải tiến và cách sử dụng
Phần lớn code bổ xung TechMaster/eris viết vào file [cuong.go](cuong.go)
Ví dụ sử dụng TechMaster/eris ở trong file [test/basic_test.go]()
### 3.1 Tạo một cảnh báo WARNING
Lỗi WARNING chỉ cần thông báo cho end user là được, không cần in ra console, không cần log ra file
Ví dụ:
- Người dùng nhập sai passwod quá 3 lần
- Đăng nhập lỗi
- Không đủ quyền truy cập
- Không tìm thấy một quyển sách người dùng mong muốn
```go
return eris.Warning("Email không hợp lệ")
```
### 3.2 Tạo một lỗi cấp độ Error
Lỗi Error là lỗi nghiệp vụ, một chu trình nào đó bị sai, cần log ra terminal, có thể ghi log file...
```go
//Tạo một lỗi, thêm HTTP status code, trả về JSON
return eris.New("Không tìm thấy bản ghi trong CSDL").StatusCode(404).EnableJSON()
```

### 3.2 Tạo System Error
System Error, lỗi hệ thống, cần in ra màn hình console và ghi ra log file. Ví System Error
- Mất kết nối tạm thời đến dịch vụ thứ 3
- Không đăng nhập được bằng Gmail hoặc GitHub
- Ổ cứng chứa ảnh đã hết chỗ, không thể upload được ảnh
- Không gọi được API Google Analytics

Việc in ra console và ghi log file để lập trình truy lại để xử lý
```go
return eris.SysError("Failed to connect Redis")
```

### 3.3 Lỗi rất nghiêm trọng `panic` không thể khôi phục, cần thoát chương trình
Với lỗi Panic cần xuất ra console, log ra file và gọi hàm panic của golang
```go
return eris.Panic("Server is down")
```

Kiểm tra xem có phải lỗi Panic không
```go
if err := connectDB(); err != nil {
	if eris.IsPanic(err) { //Hãy dùng hàm có sẵn trong eris
		//Log ra file trước rồi hãng gọi panic
		panic(err.Error())
	} else {
		return err
	}
} else {
	return nil
}
```

#### 3.4 Tạo eris từ một error khác: `New` - `NewFromMsg`
`SetType(eris.SYSERROR)` để đặt cấp độ báo lỗi
```go
if err := connectDB(connStr); err != nil {
	return eris.NewFromMsg(err, "Unable to connect DB").SetType(eris.SYSERROR)
}
```
hoặc không cần bổ xung message, đặt cấp độ WARNING
```go
if err := connectDB(connStr); err != nil {
	return eris.New(err).SetType(eris.WARNING)
}
```
#### 3.5 Đặt lại cấp độ lỗi `SetType`
```go
eris.NewFromMsg(err, "Unable to connect DB").SetType(eris.SYSERROR)
```

#### 3.6 Thêm dữ liệu để thông báo lỗi chi tiết hơn `SetData`

```go
return eris.Panic("Failed to connect to Postgresql").
		SetData(
			map[string]interface{}{
				"host": "localhost",
				"port": "5432",
			},
		)
```

Đoạn xử lý lỗi JSON sẽ như sau:
```go
switch e := err.(type) {
	case *eris.Error:
		handleErisError(e, ctx)
		if e.JSON { //Có trả về báo lỗi dạng JSON cho REST API request không?
			if e.Data == nil {
				return ctx.Status(e.Code).JSON(e.Error())
			} else {
				errorBody := map[string]interface{}{
					"error": e.Error(),
					"data":  e.Data,
				}
				return ctx.Status(e.Code).JSON(errorBody) //Trả về mô tả và thông tin bổ xung
			}
		}
	default:
	//Do other
}
```
### 4. Xử lý lỗi eris Error
#### 4.1 Kiểm tra kiểu lỗi và ép kiểu
Ứng dụng Golang có thể có nhiều loại lỗi. Do đó cần kiểm tra kiểu khi bạn làm với eris errorr.
```go
func isPanic(err error) bool {
	if e, ok := err.(*eris.Error); ok && e.ErrType == eris.PANIC {
		return true
	} else {
		return false
	}
}
```
Eris cung cấp sẵn 2 hàm kiểm tra
```go
func IsSysError(err error) bool
func IsPanic(err error) bool
```

#### 4.2 Hàm hứng lỗi cho cả ứng dụng web
Hầu hết các go web framework đều cho phép viết một hàm chung để xử lý tất cả các loại lỗi. Bạn nên tận dùng tính năng này để xử lý lỗi thay viết phải viết logic xử lý lỗi ở nhiều nơi khác nhau.
```go
// Chuyên xử lý các err mà handler trả về
func CustomErrorHandler(ctx *fiber.Ctx, err error) error {
	var statusCode = 500

	switch e := err.(type) {
	case *eris.Error:
		handleErisError(e, ctx)
		if e.JSON { //Có trả về báo lỗi dạng JSON cho REST API request không?
			if e.Data == nil {
				return ctx.Status(e.Code).JSON(e.Error())
			} else {
				errorBody := map[string]interface{}{
					"error": e.Error(),
					"data":  e.Data,
				}
				return ctx.Status(e.Code).JSON(errorBody)
			}
		}
	case *fiber.Error:
		statusCode = e.Code
		fmt.Println(err.Error())
	default:
		fmt.Println(err.Error())
	}
	//Server side error page rendering : tạo trang web báo lỗi, không áp dụng cho REST API request
	if err = ctx.Render("error/error", fiber.Map{
		"ErrorMessage": err.Error(),
		"StatusCode":   statusCode,
	}); err != nil {
		return ctx.Status(500).SendString("Internal Server Error")
	}

	return nil
}

//Hàm chuyên xử lý Eris Error có Stack Trace
func handleErisError(err *eris.Error, ctx *fiber.Ctx) {
	formattedStr := eris.ToCustomString(err, eris.StringFormat{
		Options: eris.FormatOptions{
			InvertOutput: true, // flag that inverts the error output (wrap errors shown first)
			WithTrace:    true, // flag that enables stack trace output
			InvertTrace:  true, // flag that inverts the stack trace output (top of call stack shown first)
			Top:          3,    // Giữ 3 dòng lệnh đỉnh trong Stack
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