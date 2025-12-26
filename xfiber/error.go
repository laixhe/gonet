package xfiber

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

// DefaultErrorHandler 默认错误处理
func DefaultErrorHandler() fiber.ErrorHandler {
	return func(ctx fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		var errType *fiber.Error
		switch {
		case errors.As(err, &errType):
			code = errType.Code
		default:
			err = fiber.NewError(code, err.Error())
		}
		return ctx.Status(code).JSON(err)
	}
}

// ServerError 服务器错误
func ServerError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	return fiber.NewError(fiber.StatusInternalServerError, messages[0])
}

// AuthorizedError 授权错误
func AuthorizedError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusUnauthorized)
	}
	return fiber.NewError(fiber.StatusUnauthorized, messages[0])
}

// ParamError 参数错误
func ParamError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Param Error")
	}
	return fiber.NewError(fiber.StatusUnprocessableEntity, messages[0])
}

// TipError 提示错误
func TipError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return &fiber.Error{
			Message: "Tip Error",
			Code:    427,
		}
	}
	return &fiber.Error{
		Message: messages[0],
		Code:    427,
	}
}
