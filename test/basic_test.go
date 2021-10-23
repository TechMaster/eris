package eris

import (
	"errors"
	"fmt"
	"testing"

	"github.com/TechMaster/eris"
	"github.com/stretchr/testify/assert"
)

func Test_InternalServerError(t *testing.T) {
	err := eris.New("Internal Web Server Error").InternalServerError()
	assert.Equal(t, 500, err.Code)
}

func Test_NewFrom(t *testing.T) {
	err := eris.NewFrom(errors.New("Cannot connect to database"))
	assert.Equal(t, "Cannot connect to database", err.Error())
}

func Test_NewFromMsg(t *testing.T) {
	err := eris.NewFromMsg(errors.New("Cannot connect to database"), "Lỗi kết nối CSDL")
	assert.Equal(t, "Cannot connect to database : Lỗi kết nối CSDL", err.Error())
}

func Test_Warning(t *testing.T) {
	err := eris.Warning("Invalid Email")
	assert.Equal(t, eris.WARNING, err.ErrType)
}

func Test_BadRequest(t *testing.T) {
	err := eris.New("Phone is invalid").BadRequest()
	assert.Equal(t, 400, err.Code)
}

func Test_SysError(t *testing.T) {
	err := eris.SysError("Cannot connect to Redis")
	assert.Equal(t, eris.SYSERROR, err.ErrType)
	assert.True(t, eris.IsSysError(err))
}
func foo() error {
	return eris.New("foo")
}
func bar() error {
	return foo()
}
func rock() error {
	return bar()
}
func Test_stack_trace(t *testing.T) {
	if eris_err, ok := rock().(*eris.Error); ok {
		eris_string_format := eris.StringFormat{
			Options: eris.FormatOptions{
				InvertOutput: false, // flag that inverts the error output (wrap errors shown first)
				WithTrace:    true,  // flag that enables stack trace output
				InvertTrace:  true,  // flag that inverts the stack trace output (top of call stack shown first)
				WithExternal: false,
				Top:          3, // Chỉ lấy 3 dòng lệnh đầu tiên
			},
			MsgStackSep:  "\n",  // separator between error messages and stack frame data
			PreStackSep:  "\t",  // separator at the beginning of each stack frame
			StackElemSep: " | ", // separator between elements of each stack frame
			ErrorSep:     "\n",  // separator between each error in the chain
		}
		formattedStr := eris.ToCustomString(eris_err, eris_string_format)
		fmt.Println(formattedStr)
		assert.Contains(t, formattedStr, "test.foo")
		assert.Contains(t, formattedStr, "test.bar")
		assert.Contains(t, formattedStr, "test.rock")
		assert.Contains(t, formattedStr, "basic_test.go")
	} else {
		assert.FailNow(t, "not eris error")
	}
}

func Test_fluent_call(t *testing.T) {
	err := eris.NewFrom(errors.New("Cannot connect to database")).
		InternalServerError().
		SetType(eris.PANIC).
		SetData(map[string]interface{}{
			"host": "localhost",
			"port": 5432,
			"user": "cuong",
		})

	assert.Equal(t, err.ErrType, eris.PANIC)
	assert.Equal(t, 500, err.Code)
	assert.Equal(t, "localhost", err.Data["host"])
	assert.Equal(t, 5432, err.Data["port"])
	assert.Equal(t, "cuong", err.Data["user"])
}
