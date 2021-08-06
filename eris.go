package eris

import (
	"fmt"
	"io"
	"reflect"
)

//Cuong bổ xung thêm chức năng báo cấp độ lỗi
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
	JSON    bool      // true if error will resonse as JSON for REST request, false to render error page at server side
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
	return New(err.Error())
}

//Bao lấy một error và thêm báo lỗi
func NewFromMsg(err error, msg string) *Error {
	return New(err.Error() + " : " + msg)
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

//Kiểm tra xem có phải là lỗi hệ thống
func (error *Error) IsSysError() bool {
	return error.ErrType == SYSERROR
}

//Kiểm tra xem có phải là lỗi nghiêm trọng
func (error *Error) IsPanic() bool {
	return error.ErrType == PANIC
}

//----- Hết đoạn code Cường bổ xung

// Errorf creates a new root error with a formatted message.
func Errorf(format string, args ...interface{}) error {
	stack := callers(3)
	return &Error{
		global: stack.isGlobal(),
		msg:    fmt.Sprintf(format, args...),
		stack:  stack,
	}
}

// Wrap adds additional context to all error types while maintaining the type of the original error.
//
// This method behaves differently for each error type. For root errors, the stack trace is reset to the current
// callers which ensures traces are correct when using global/sentinel error values. Wrapped error types are simply
// wrapped with the new context. For external types (i.e. something other than root or wrap errors), this method
// attempts to unwrap them while building a new error chain. If an external type does not implement the unwrap
// interface, it flattens the error and creates a new root error from it before wrapping with the additional
// context.
func Wrap(err error, msg string) error {
	return wrap(err, fmt.Sprint(msg))
}

// Wrapf adds additional context to all error types while maintaining the type of the original error.
//
// This is a convenience method for wrapping errors with formatted messages and is otherwise the same as Wrap.
func Wrapf(err error, format string, args ...interface{}) error {
	return wrap(err, fmt.Sprintf(format, args...))
}

func wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	// callers(4) skips runtime.Callers, stack.callers, this method, and Wrap(f)
	stack := callers(4)
	// caller(3) skips stack.caller, this method, and Wrap(f)
	// caller(skip) has a slightly different meaning which is why it's not 4 as above
	frame := caller(3)
	switch e := err.(type) {
	case *Error:
		if e.global {
			// create a new root error for global values to make sure nothing interferes with the stack
			err = &Error{
				global: e.global,
				msg:    e.msg,
				stack:  stack,
			}
		} else {
			// insert the frame into the stack
			e.stack.insertPC(*stack)
		}
	case *wrapError:
		// insert the frame into the stack
		if root, ok := Cause(err).(*Error); ok {
			root.stack.insertPC(*stack)
		}
	default:
		// return a new root error that wraps the external error
		return &Error{
			msg:   msg,
			ext:   e,
			stack: stack,
		}
	}

	return &wrapError{
		msg:   msg,
		err:   err,
		frame: frame,
	}
}

// Unwrap returns the result of calling the Unwrap method on err, if err's type contains an Unwrap method
// returning error. Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}

// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained by repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if it implements a method
// Is(error) bool such that Is(target) returns true.
func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflect.TypeOf(target).Comparable()
	for {
		if isComparable && err == target {
			return true
		}
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

// As finds the first error in err's chain that matches target. If there's a match, it sets target to that error
// value and returns true. Otherwise, it returns false.
//
// The chain consists of err itself followed by the sequence of errors obtained by repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value pointed to by target,
// or if the error has a method As(interface{}) bool such that As(target) returns true.
func As(err error, target interface{}) bool {
	if target == nil || err == nil {
		return false
	}
	val := reflect.ValueOf(target)
	typ := val.Type()

	// target must be a non-nil pointer
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		return false
	}

	// *target must be interface or implement error
	if e := typ.Elem(); e.Kind() != reflect.Interface && !e.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return false
	}

	for {
		errType := reflect.TypeOf(err)
		if errType != reflect.TypeOf(&wrapError{}) && errType != reflect.TypeOf(&Error{}) && reflect.TypeOf(err).AssignableTo(typ.Elem()) {
			val.Elem().Set(reflect.ValueOf(err))
			return true
		}
		if x, ok := err.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

// Cause returns the root cause of the error, which is defined as the first error in the chain. The original
// error is returned if it does not implement `Unwrap() error` and nil is returned if the error is nil.
func Cause(err error) error {
	for {
		uerr := Unwrap(err)
		if uerr == nil {
			return err
		}
		err = uerr
	}
}

// StackFrames returns the trace of an error in the form of a program counter slice.
// Use this method if you want to pass the eris stack trace to some other error tracing library.
func StackFrames(err error) []uintptr {
	for err != nil {
		switch err := err.(type) {
		case *Error:
			return err.StackFrames()
		case *wrapError:
			return err.StackFrames()
		default:
			return []uintptr{}
		}
	}
	return []uintptr{}
}

func (e *Error) Error() string {
	return fmt.Sprint(e)
}

func (e *Error) Format(s fmt.State, verb rune) {
	printError(e, s, verb)
}

func (e *Error) Is(target error) bool {
	if err, ok := target.(*Error); ok {
		return e.msg == err.msg
	}
	return e.msg == target.Error()
}

func (e *Error) As(target interface{}) bool {
	t := reflect.Indirect(reflect.ValueOf(target)).Interface()
	if err, ok := t.(*Error); ok {
		if e.msg == err.msg {
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(err))
			return true
		}
	}
	return false
}

func (e *Error) Unwrap() error {
	return e.ext
}

// StackFrames returns the trace of a root error in the form of a program counter slice.
// This method is currently called by an external error tracing library (Sentry).
func (e *Error) StackFrames() []uintptr {
	return *e.stack
}

type wrapError struct {
	msg   string // wrap error message
	err   error  // error type representing the next error in the chain
	frame *frame // wrap error stack frame
}

func (e *wrapError) Error() string {
	return fmt.Sprint(e)
}

func (e *wrapError) Format(s fmt.State, verb rune) {
	printError(e, s, verb)
}

func (e *wrapError) Is(target error) bool {
	if err, ok := target.(*wrapError); ok {
		return e.msg == err.msg
	}
	return e.msg == target.Error()
}

func (e *wrapError) As(target interface{}) bool {
	t := reflect.Indirect(reflect.ValueOf(target)).Interface()
	if err, ok := t.(*wrapError); ok {
		if e.msg == err.msg {
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(err))
			return true
		}
	}
	return false
}

func (e *wrapError) Unwrap() error {
	return e.err
}

// StackFrames returns the trace of a wrap error in the form of a program counter slice.
// This method is currently called by an external error tracing library (Sentry).
func (e *wrapError) StackFrames() []uintptr {
	return []uintptr{e.frame.pc()}
}

func printError(err error, s fmt.State, verb rune) {
	var withTrace bool
	switch verb {
	case 'v':
		if s.Flag('+') {
			withTrace = true
		}
	}
	str := ToString(err, withTrace)
	_, _ = io.WriteString(s, str)
}
