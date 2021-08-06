# Cải tiến từ thư viện [https://github.com/rotisserie/eris](https://github.com/rotisserie/eris)

**Whole source code of this package is credited to rotisserie/eris.**

Ưu điểm lớn nhất của [rotisserie/eris](https://github.com/rotisserie/eris) đó là lỗi bao gồm cả stack trace giúp lập trình viên nhanh chóng tìm lỗi. Những gì tôi bổ xung thêm để ở file [cuong.go](cuong.go)

## Hướng dẫn sử dụng
### 1. Cài đặt package
```
go get -u github.com/TechMaster/eris
```

### 2. Sử dụng eris
#### 2.0 Tạo một cảnh báo WARNING
Lỗi WARNING chỉ cần thông báo cho end user là được, không cần in ra console, không cần log ra file
Ví dụ:
- Người dùng nhập sai passwod quá 3 lần
- Đăng nhập lỗi
- Không đủ quyền truy cập
- Không tìm thấy một quyển sách người dùng mong muốn
```go
return eris.Warning("Không tìm thấy sách trong CSDL")
```
#### 2.1 Tạo một lỗi cấp độ Error
Lỗi Error là lỗi nghiệp vụ, một chu trình nào đó bị sai, cần log ra terminal, có thể ghi log file...
```go
//Tạo một lỗi, thêm HTTP status code, trả về JSON
return eris.New("Không tìm thấy bản ghi trong CSDL").StatusCode(404).EnableJSON()
```

#### 2.2 Tạo System Error
System Error, lỗi hệ thống, cần in ra màn hình console và ghi ra log file. Ví System Error
- Mất kết nối tạm thời đến dịch vụ thứ 3
- Không đăng nhập được bằng Gmail hoặc GitHub
- Ổ cứng chứa ảnh đã hết chỗ, không thể upload được ảnh
- Không gọi được API Google Analytics

Việc in ra console và ghi log file để lập trình truy lại để xử lý
```go
return eris.SysError("Failed to connect Redis")
```

#### 2.3 Lỗi rất rất nghiêm trọng, hệ thống sập ngay tức thì
Với lỗi Panic cần xuất ra console, log ra file. Nếu không ```EnableJSON()``` có nghĩa lỗi này sẽ được trả về trang báo lỗi server side rendering error page.
```go
return eris.Panic("Server is down")
```
#### 2.4 Tạo eris từ một error khác
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
#### 2.5 Lỗi phải trả về JSON bằng `.EnableJSON()`
```go
return eris.Warning("Không tìm được sách").StatusCode(404).EnableJSON()
```

#### 2.6 Đặt lại cấp độ lỗi
```go
eris.NewFromMsg(err, "Unable to connect DB").SetType(eris.SYSERROR)
```

#### 2.7 Thêm dữ liệu để thông báo lỗi chi tiết hơn
```go
data := map[string]interface{}{
	"host":  "192.168.1.1",
	"port":  8008,
	"roles": []string{"admin", "editor", "user"},
}

return eris.New("Unable connect to login").SetData(data).StatusCode(fiber.StatusUnauthorized).EnableJSON()
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





### 3. Xử lý lỗi eris Error
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