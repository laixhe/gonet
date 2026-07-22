package xfiber

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

// DefaultErrorHandler 默认错误处理，将 error 转换为 JSON 响应。
// *fiber.Error 类型使用其内置的 HTTP 状态码，其余类型返回 500。
func DefaultErrorHandler() fiber.ErrorHandler {
	return func(ctx fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		var errType *fiber.Error
		if errors.As(err, &errType) {
			code = errType.Code
		} else {
			err = fiber.NewError(code, err.Error())
		}
		return ctx.Status(code).JSON(err)
	}
}

// ServerError 创建 500 服务器内部错误。
// 不传 messages 时不带自定义消息，由 DefaultErrorHandler 合成错误文本。
func ServerError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	return fiber.NewError(fiber.StatusInternalServerError, messages[0])
}

// AuthorizedError 创建 401 授权错误。
// 不传 messages 时不带自定义消息，由 DefaultErrorHandler 合成错误文本。
func AuthorizedError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusUnauthorized)
	}
	return fiber.NewError(fiber.StatusUnauthorized, messages[0])
}

// ParamError 创建 422 参数校验错误。
// 适用于请求参数不合法（如缺少必填字段、格式错误）。
// 不传 messages 时使用默认消息 "Param Error"。
func ParamError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Param Error")
	}
	return fiber.NewError(fiber.StatusUnprocessableEntity, messages[0])
}

// TimeoutError 创建 408 请求超时错误。
// 不传 messages 时使用默认消息 "Request Timeout"。
func TimeoutError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusRequestTimeout, "Request Timeout")
	}
	return fiber.NewError(fiber.StatusRequestTimeout, messages[0])
}

// NotFoundError 创建 404 未找到错误。
// 不传 messages 时使用默认消息 "Not Found"。
func NotFoundError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Not Found")
	}
	return fiber.NewError(fiber.StatusNotFound, messages[0])
}

// RateLimitError 创建 429 请求过多错误。
// 不传 messages 时使用默认消息 "Too Many Requests"。
func RateLimitError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusTooManyRequests, "Too Many Requests")
	}
	return fiber.NewError(fiber.StatusTooManyRequests, messages[0])
}

// TipError 创建 400 业务提示错误。
// 适用于向客户端返回可直接展示的提示（如"用户名已存在"、"验证码错误"）。
// 不传 messages 时使用默认消息 "Tip Error"。
func TipError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Tip Error")
	}
	return fiber.NewError(fiber.StatusBadRequest, messages[0])
}
