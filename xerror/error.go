package xerror

import (
	"errors"
)

type IError interface {
	HttpStatus() int
	Code() int
	Msg() string
	Error() string
}

type Error struct {
	httpStatus int
	code       int
	msg        string
	err        error
}

func (e *Error) HttpStatus() int {
	return e.httpStatus
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Msg() string {
	return e.msg
}

func (e *Error) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func NewError(httpStatus int, code int, msg string, err error) *Error {
	return &Error{
		httpStatus: httpStatus,
		code:       code,
		msg:        msg,
		err:        err,
	}
}

func NewErrorStr(httpStatus int, code int, msg string, errStr string) *Error {
	return &Error{
		httpStatus: httpStatus,
		code:       code,
		msg:        msg,
		err:        errors.New(errStr),
	}
}
