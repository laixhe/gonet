package xresponse

import (
	"github.com/laixhe/gonet/xerror"
)

// ResponseModel 响应请求的公共模型
type ResponseModel struct {
	Code int32  `json:"code"`           // 响应码
	Msg  string `json:"msg"`            // 响应信息
	Data any    `json:"data,omitempty"` // 数据
}

func Error(err xerror.IError) *ResponseModel {
	return &ResponseModel{
		Code: err.Code(),
		Msg:  err.Error(),
	}
}
