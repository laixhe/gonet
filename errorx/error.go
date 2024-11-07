package errorx

import (
	"errors"

	"github.com/laixhe/gonet/proto/gen/ecode"
)

type Error struct {
	Code ecode.ECode
	Err  error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func NewError(code ecode.ECode, err error) *Error {
	return &Error{
		Code: code,
		Err:  err,
	}
}

func ServiceError(err error) *Error {
	return &Error{
		Code: ecode.ECode_Service,
		Err:  err,
	}
}

func ServiceErrorStr(err string) *Error {
	return &Error{
		Code: ecode.ECode_Service,
		Err:  errors.New(err),
	}
}

func ParamError(err error) *Error {
	return &Error{
		Code: ecode.ECode_Param,
		Err:  err,
	}
}

func ParamErrorStr(err string) *Error {
	return &Error{
		Code: ecode.ECode_Param,
		Err:  errors.New(err),
	}
}

func TipMessageError(err error) *Error {
	return &Error{
		Code: ecode.ECode_TipMessage,
		Err:  err,
	}
}

func TipMessageErrorStr(err string) *Error {
	return &Error{
		Code: ecode.ECode_TipMessage,
		Err:  errors.New(err),
	}
}

func RepeatError(err error) *Error {
	return &Error{
		Code: ecode.ECode_Repeat,
		Err:  err,
	}
}

func RepeatErrorStr(err string) *Error {
	return &Error{
		Code: ecode.ECode_Repeat,
		Err:  errors.New(err),
	}
}

func AuthInvalidError(err error) *Error {
	return &Error{
		Code: ecode.ECode_AuthInvalid,
		Err:  err,
	}
}

func AuthInvalidErrorStr(err string) *Error {
	return &Error{
		Code: ecode.ECode_AuthInvalid,
		Err:  errors.New(err),
	}
}

func New(code int32, err error) *Error {
	return &Error{
		Code: ecode.ECode(code),
		Err:  err,
	}
}

func NewStr(code int32, err string) *Error {
	return &Error{
		Code: ecode.ECode(code),
		Err:  errors.New(err),
	}
}
