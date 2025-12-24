package xfiber

import "github.com/gofiber/fiber/v3"

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
		return fiber.NewError(fiber.StatusUnprocessableEntity, "参数错误")
	}
	return fiber.NewError(fiber.StatusUnprocessableEntity, messages[0])
}

// TipError 提示错误
func TipError(messages ...string) *fiber.Error {
	if len(messages) == 0 {
		return &fiber.Error{
			Message: "提示错误",
			Code:    427,
		}
	}
	return &fiber.Error{
		Message: messages[0],
		Code:    427,
	}
}
