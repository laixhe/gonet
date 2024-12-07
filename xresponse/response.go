package xresponse

import (
	"github.com/laixhe/gonet/xerror"
)

// ResponseModel 响应模型
type ResponseModel struct {
	Code int32  `json:"code"`           // 响应码
	Msg  string `json:"msg"`            // 响应错误信息
	Data any    `json:"data,omitempty"` // 数据
}

// Error 响应错误
func Error(err xerror.IError) *ResponseModel {
	return &ResponseModel{
		Code: err.Code(),
		Msg:  err.Error(),
	}
}
