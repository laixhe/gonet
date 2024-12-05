package xerror

type IError interface {
	Code() int32
	Error() string
}

type Error struct {
	code int32
	err  string
}

func (e *Error) Code() int32 {
	return e.code
}

func (e *Error) Error() string {
	return e.err
}

func New(code int32, err error) Error {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	return Error{
		code: code,
		err:  errStr,
	}
}

func NewStr(code int32, errStr string) Error {
	return Error{
		code: code,
		err:  errStr,
	}
}
