# Cải tiến từ thư viện [https://github.com/rotisserie/eris](https://github.com/rotisserie/eris)

## Ưu điểm bắt lỗi, xử lý lỗi với [TechMaster/eris](https://github.com/TechMaster/eris)
**Noticed: I customized code from rotisserie/eris.**

Ưu điểm lớn nhất của [rotisserie/eris](https://github.com/rotisserie/eris) đó là lỗi bao gồm cả stack trace giúp lập trình viên nhanh chóng tìm lỗi. Những tính năng bổ xung viết ở [cuong.go](cuong.go). Ví dụ chi tiết sử dụng TechMaster/eris các trường hợp ở đây [demofiber](https://github.com/TechMaster/demofiber)

1. Thêm các trường để lập trình bổ xung 
	```go
	ErrType ErrorType  //Loại lỗi: WARNING, ERROR, SYSERROR, PANIC
	Code    int        //HTTP Status code
	JSON    bool       //true trả JSON cho REST request, false hiển thị trang báo lỗi
	Data    map[string]interface{} //Chứa thông tin bổ xung
	```
2. Tạo lỗi bằng nỗi chuỗi các hàm

	```go
	return new eris.SysError("Cannot connect DB").StatusCode(500).EnableJSON().SetData(
		map[string]interface{}{
				"host": "localhost",
				"port": "5432",
				"db"  : "inventory",
		},
	)
	```
3. Bao lấy lỗi thông thường khác bằng `eris.NewFrom` và `eris.NewFromMsg`
4. Kiểm tra mức độ lỗi bằng `eris.IsSysErr` và `eris.IsPanic`
5. Cấu hình để loại bỏ bớt vài hàm dưới cùng StackTrace. Xem [format.go](format.go)
   ```go
	 type FormatOptions struct {
		...
		Skip int  //Số hàm dưới cùng stack trace sẽ bỏ qua không cần in
	}
	```


## Hướng dẫn sử dụng eris
### 1. Cài đặt package
```
go get -u github.com/TechMaster/eris
```

### 2. Tạo lỗi - bắt lỗi - xử lý lỗi - báo  lỗi - log lỗi
Lập trình viên Golang cần chú ý
1. Golang không có try catch exception, chỉ có hàm trả về lỗi
2. Một khi đã viết hàm Go bạn cần phải quyết định hàm này có trả về lỗi hay không?  90% hàm phải trả về lỗi
3. Lỗi được tạo ra để xử lý và cần phải được xử lý lỗi đến nơi đến trốn. Tuyệt đối không được dập lỗi, lờ đi.
4. Nếu một hàm trả về nhiều tham số, thì tham số lỗi luôn để cuối cùng
	```go
	func Foo() (result string, count int, err error)
	```
4. Bất kỳ lỗi nào trong Golang đều phải tuân thủ interface
	```go
	type error interface {
    	Error() string
	}	
	```

#### 2.1 Các bước làm việc với lỗi
1. Define function return error: Định nghĩa hàm trả về lỗi
2. Create error: Tạo error phù hợp
3. Handle error: Xử lý lỗi gồm có kiểm tra loại lỗi, mức độ lỗi, mã lỗi
4. Report error: Báo lỗi cho client, cần quyết định nội dung chi tiết đến mức nào và kiểu báo lỗi. Có hai kiểu báo lỗi:
	- Ứng dụng Server Side Rendering thì trả về trang báo lỗi error page
	- Ứng dụng Client Side, Mobile thì trả về JSON error cùng HTTP status code 
5. Log error: Lỗi nghiêm trọng cần được in ra màn hình console và ghi ra file. Với lỗi panic bắt buộc dừng chương trình bằng hàm `panic("error message")`


#### 2.2 Căn bản về lỗi
Một lỗi đầy đủ cần có:
1. Mô tả lỗi
2. Cấp độ lỗi: WARNING, ERROR, SYSERROR, PANIC quyết định cách thức dev báo cáo lỗi và log lỗi
3. Stack Trace danh sách các hàm gọi nhau gây ra lỗi
4. HTTP Status Code nếu là lỗi sẽ trả về cho REST Client
5. Dữ liệu bổ trợ cho lỗi

Những hành động của lập trình với lỗi:
1. Báo cáo lỗi cho client: trả về trang báo lỗi dễ hiểu, thân thiện
2. Trả về lỗi dạng JSON đối với REST API request
3. In lỗi ra màn hình terminal, sẽ bị mất khi docker container nâng cấp
4. Ghi lỗi vào file, bền vững hơn
5. Bỏ qua lỗi nếu thấy cần (hãn hữu thôi nhé)
6. Nâng cấp độ lỗi lên mức cao hơn
7. Tạo ra một lỗi từ một lỗi khác để thêm thông báo, và dữ liệu bổ trợ

Không xử lý lỗi đúng dẫn đến vấn đề gì?
1. Người dùng không hiểu chuyện gì đã xảy ra
2. Lập trình viên không dò vết (không xem được Stack Trace của lỗi), vì lỗi qua chung chung, khó hiểu
3. Hệ thống sập vì lỗi không được xử lý đúng, chương trình chạy tiếp với biến rỗng (nil)

#### 2.3 Log lỗi
Cần phân biệt rõ báo lỗi và log lỗi. Báo lỗi dùng để báo cho client, người dùng cuối. Còn log lỗi là cho hệ thống nội bộ và lập trình viên debug, fix lỗi. Do đó Log lỗi phải chi tiết đầy đủ, cần gồm cả stack trace và thông tin hoàn cảnh lỗi phát sinh. Ngược lại báo lỗi cần ưu tiên sự thân thiện với người dùng.

Log lỗi sẽ có 2 cấp độ:
1. In ra màn hình console
2. Ghi vào file log

Bạn có thể sử dụng các hàm thông thường của Golang hay một thư viện như Uber Zap để log lỗi. Tránh viết logic log lỗi ở mọi nơi khiến code vừa dài, mà vừa phụ thuộc chặt (tightly coupling) vào một thư viện báo lỗi bên thứ ba. Nên tận dụng một hàm xử lý lỗi chung gắn với ứng dụng. Muốn được như vậy, ta phải tuần tự trả về lỗi từ hàm con ra hàm cha, từ hàm cha ra hàm ông, cụ, kỵ...
### 3. Sử dụng eris
#### 3.0 Tạo một cảnh báo WARNING
Lỗi WARNING chỉ cần thông báo cho end user là được, không cần in ra console, không cần log ra file
Ví dụ:
- Người dùng nhập sai passwod quá 3 lần
- Đăng nhập lỗi
- Không đủ quyền truy cập
- Không tìm thấy một quyển sách người dùng mong muốn
```go
return eris.Warning("Email không hợp lệ")
```
#### 3.1 Tạo một lỗi cấp độ Error
Lỗi Error là lỗi nghiệp vụ, một chu trình nào đó bị sai, cần log ra terminal, có thể ghi log file...
```go
//Tạo một lỗi, thêm HTTP status code, trả về JSON
return eris.New("Không tìm thấy bản ghi trong CSDL").StatusCode(404).EnableJSON()
```

#### 3.2 Tạo System Error
System Error, lỗi hệ thống, cần in ra màn hình console và ghi ra log file. Ví System Error
- Mất kết nối tạm thời đến dịch vụ thứ 3
- Không đăng nhập được bằng Gmail hoặc GitHub
- Ổ cứng chứa ảnh đã hết chỗ, không thể upload được ảnh
- Không gọi được API Google Analytics

Việc in ra console và ghi log file để lập trình truy lại để xử lý
```go
return eris.SysError("Failed to connect Redis")
```

#### 3.3 Lỗi rất nghiêm trọng `panic`
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
#### 3.5 Lỗi phải trả về JSON bằng `.EnableJSON()`
```go
return eris.Warning("Không tìm được sách").StatusCode(404).EnableJSON()
```

#### 3.6 Đặt lại cấp độ lỗi `SetType`
```go
eris.NewFromMsg(err, "Unable to connect DB").SetType(eris.SYSERROR)
```

#### 3.7 Thêm dữ liệu để thông báo lỗi chi tiết hơn `SetData`

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