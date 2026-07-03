package douyin

import "fmt"

type ErrorData struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewErrorData(code int, msg string) *ErrorData {
	return &ErrorData{
		Code: code,
		Msg:  msg,
	}
}

func (e *ErrorData) Error() string {
	return fmt.Sprintf("bouyin(%d:%s)", e.Code, e.Msg)
}
