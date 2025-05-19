package xgin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"

	"github.com/laixhe/gonet/xerror"
)

// ResponseModel 响应模型
type ResponseModel struct {
	Code int    `json:"code"`           // 响应码
	Msg  string `json:"msg"`            // 响应错误信息
	Data any    `json:"data,omitempty"` // 数据
}

func NewResponseModelError(err xerror.IError) *ResponseModel {
	return &ResponseModel{
		Code: err.Code(),
		Msg:  err.Msg(),
		Data: err.Error(),
	}
}

// Success 成功
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// SuccessResponse 成功
func SuccessResponse(c *gin.Context, data any) {
	c.JSON(http.StatusOK, &ResponseModel{
		Data: data,
	})
}

// ErrorResponse 错误
func ErrorResponse(c *gin.Context, errorAny any) {
	switch errorAny.(type) {
	case xerror.IError:
		err := errorAny.(xerror.IError)
		c.JSON(err.HttpStatus(), ResponseModel{
			Code: err.Code(),
			Msg:  err.Msg(),
			Data: err.Error(),
		})
	case validator.ValidationErrors:
		err := IErrorParse(errors.New(TranslatorErrorString(errorAny.(validator.ValidationErrors))))
		c.JSON(err.HttpStatus(), ResponseModel{
			Code: err.Code(),
			Msg:  err.Msg(),
			Data: err.Error(),
		})
	default:
		c.JSON(http.StatusInternalServerError, ResponseModel{
			Code: http.StatusInternalServerError,
			Msg:  "服务器异常",
			Data: errorAny,
		})
	}
}

func IErrorServer(err error) xerror.IError {
	return xerror.NewError(http.StatusInternalServerError, http.StatusInternalServerError, "服务器异常", err)
}

func IErrorAuthInvalid(err error) xerror.IError {
	return xerror.NewError(http.StatusUnauthorized, http.StatusUnauthorized, "未授权", err)
}

func IErrorParse(err error) xerror.IError {
	return xerror.NewError(http.StatusUnprocessableEntity, http.StatusUnprocessableEntity, "参数错误", err)
}
