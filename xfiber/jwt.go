package xfiber

import (
	contribJwt "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

// UseJwt 中间件-JWT
func UseJwt(config ...contribJwt.Config) fiber.Handler {
	if len(config) == 0 {
		config = []contribJwt.Config{{}}
	}
	if config[0].ErrorHandler == nil {
		config[0].ErrorHandler = JwtErrorHandler
	}
	return contribJwt.New(config[0])
}

// JwtErrorHandler 自定义JWT错误处理
func JwtErrorHandler(ctx fiber.Ctx, err error) error {
	log.WithContext(ctx.Context()).
		Errorf("jwt: %s error: %v", ctx.Get(fiber.HeaderAuthorization), err)
	return ctx.Status(fiber.StatusUnauthorized).JSON(AuthorizedError())
}

// JwtErrorHandlerNext 自定义JWT错误处理
func JwtErrorHandlerNext(ctx fiber.Ctx, err error) error {
	log.WithContext(ctx.Context()).
		Errorf("jwt: %s error: %v", ctx.Get(fiber.HeaderAuthorization), err)
	return ctx.Next()
}
