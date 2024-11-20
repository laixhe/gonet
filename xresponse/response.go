package xresponse

import (
	"encoding/json"
	"errors"

	validator "github.com/go-playground/validator/v10"

	"github.com/laixhe/gonet/proto/gen/ecode"
	"github.com/laixhe/gonet/xerror"
	"github.com/laixhe/gonet/xgin/xvalidator"
)

// ResponseModel 响应请求的公共模型
type ResponseModel struct {
	Code int32  `json:"code"`           // 响应码
	Msg  string `json:"msg"`            // 响应信息
	Data any    `json:"data,omitempty"` // 数据
}

func ResponseError(err error) *ResponseModel {
	var ex *xerror.Error
	if errors.As(err, &ex) {
		return &ResponseModel{
			Code: ex.Code,
			Msg:  ex.Error(),
		}
	}
	var ev validator.ValidationErrors
	if errors.As(err, &ev) {
		return &ResponseModel{
			Code: int32(ecode.ECode_Param),
			Msg:  xvalidator.TranslatorErrorString(ev),
		}
	}
	var ejut *json.UnmarshalTypeError
	if errors.As(err, &ejut) {
		return &ResponseModel{
			Code: int32(ecode.ECode_Param),
			Msg:  ejut.Error(),
		}
	}
	var ejse *json.SyntaxError
	if errors.As(err, &ejse) {
		return &ResponseModel{
			Code: int32(ecode.ECode_Param),
			Msg:  ejse.Error(),
		}
	}
	return &ResponseModel{
		Code: int32(ecode.ECode_Service),
		Msg:  err.Error(),
	}
}
