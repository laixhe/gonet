package xgin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorRecoveryFunc 默认 panic 恢复处理函数，返回 500 JSON 响应。
var ErrorRecoveryFunc = func(ctx *gin.Context, err any) {
	ctx.JSON(http.StatusInternalServerError, ServerError())
}

// Error 统一错误响应结构，实现 error 接口，可直接作为 HTTP JSON 响应体和 Go error 使用。
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error 实现 error 接口，返回错误消息文本。
func (e *Error) Error() string {
	return e.Message
}

// ServerError 创建 500 服务器内部错误。
// 不传 messages 时使用默认消息 "Internal Server Error"。
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

// AuthorizedError 创建 401 授权错误。
// 不传 messages 时使用默认消息 "Unauthorized"。
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

// ParamError 创建 422 参数校验错误。
// 适用于请求参数不合法（如缺少必填字段、格式错误）。
// 不传 messages 时使用默认消息 "Param Error"。
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

// TipError 创建 400 业务提示错误，用于向客户端返回可展示的提示信息（如"余额不足"、"操作太频繁"等）。
// 不传 messages 时使用默认消息 "Tip Error"。
func TipError(messages ...string) *Error {
	if len(messages) == 0 {
		return &Error{
			Code:    http.StatusBadRequest,
			Message: "Tip Error",
		}
	}
	return &Error{
		Code:    http.StatusBadRequest,
		Message: messages[0],
	}
}

// TimeoutError 创建 408 请求超时错误。
// 不传 messages 时使用默认消息 "Request Timeout"。
func TimeoutError(messages ...string) *Error {
	if len(messages) == 0 {
		return &Error{
			Code:    http.StatusRequestTimeout,
			Message: "Request Timeout",
		}
	}
	return &Error{
		Code:    http.StatusRequestTimeout,
		Message: messages[0],
	}
}

// NotFoundError 创建 404 未找到错误。
// 不传 messages 时使用默认消息 "Not Found"。
func NotFoundError(messages ...string) *Error {
	if len(messages) == 0 {
		return &Error{
			Code:    http.StatusNotFound,
			Message: "Not Found",
		}
	}
	return &Error{
		Code:    http.StatusNotFound,
		Message: messages[0],
	}
}

// RateLimitError 创建 429 请求过多错误。
// 不传 messages 时使用默认消息 "Too Many Requests"。
func RateLimitError(messages ...string) *Error {
	if len(messages) == 0 {
		return &Error{
			Code:    http.StatusTooManyRequests,
			Message: "Too Many Requests",
		}
	}
	return &Error{
		Code:    http.StatusTooManyRequests,
		Message: messages[0],
	}
}
