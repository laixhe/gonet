package xfiber

import (
	contribJwt "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

// UseJwt 创建 JWT 鉴权中间件。
// 必须传入 SigningKey 配置签名密钥；不传 ErrorHandler 时默认返回 401 JSON。
//
// 用法示例：
//
//	app.Use(UseJwt(contribJwt.Config{
//	    SigningKey: contribJwt.SigningKey{Key: []byte("secret"), JWTAlg: "HS256"},
//	}))
func UseJwt(config ...contribJwt.Config) fiber.Handler {
	if len(config) == 0 {
		config = []contribJwt.Config{{}}
	}
	if config[0].ErrorHandler == nil {
		config[0].ErrorHandler = JwtErrorHandler
	}
	return contribJwt.New(config[0])
}

// JwtErrorHandler JWT 校验失败时记录错误并返回 401 响应。
func JwtErrorHandler(ctx fiber.Ctx, err error) error {
	log.WithContext(ctx.Context()).
		Errorf("jwt: %s error: %v", maskAuth(ctx.Get(fiber.HeaderAuthorization)), err)
	return ctx.Status(fiber.StatusUnauthorized).JSON(AuthorizedError())
}

// JwtOptionalErrorHandler JWT 可选模式错误处理：校验失败记录日志但不拦截请求。
// 适用于同时支持登录用户和匿名用户访问的接口（如文章详情页）。
func JwtOptionalErrorHandler(ctx fiber.Ctx, err error) error {
	log.WithContext(ctx.Context()).
		Errorf("jwt: %s error: %v", maskAuth(ctx.Get(fiber.HeaderAuthorization)), err)
	return ctx.Next()
}
