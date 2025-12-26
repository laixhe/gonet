package xgin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorRecoveryFunc 错误恢复函数
var ErrorRecoveryFunc = func(ctx *gin.Context, err any) {
	ctx.JSON(http.StatusInternalServerError, ServerError())
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

// ServerError 服务器错误
func ServerError(messages ...string) *Error {
	if len(messages) == 0 {
		return &Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal Server Error",
		}
	}
	return &Error{
		Code:    http.StatusInternalServerError,
		Message: messages[0],
	}
}

// AuthorizedError 授权错误
func AuthorizedError(messages ...string) *Error {
	if len(messages) == 0 {
		return &Error{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}
	return &Error{
		Code:    http.StatusUnauthorized,
		Message: messages[0],
	}
}

// ParamError 参数错误
func ParamError(messages ...string) *Error {
	if len(messages) == 0 {
		return &Error{
			Code:    http.StatusUnprocessableEntity,
			Message: "Param Error",
		}
	}
	return &Error{
		Code:    http.StatusUnprocessableEntity,
		Message: messages[0],
	}
}

// TipError 提示错误
func TipError(messages ...string) *Error {
	if len(messages) == 0 {
		return &Error{
			Code:    427,
			Message: "Tip Error",
		}
	}
	return &Error{
		Code:    427,
		Message: messages[0],
	}
}
