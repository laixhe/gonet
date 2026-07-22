package xfiber

import (
	"github.com/gofiber/fiber/v3"
)

// NotFound 返回 404 兜底处理器，未匹配路由时返回 JSON 格式错误响应。
// 该处理器不调用 c.Next()，命中后直接终止请求链。
func NotFound() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		return ctx.Status(fiber.StatusNotFound).JSON(NotFoundError())
	}
}

// UseNotFound 注册 404 兜底中间件，不传 handler 时使用默认 NotFound()。
// 必须在所有路由定义之后调用，利用 fiber app.Use 的栈顺序作为兜底。
//
// 典型用法：
//
//	s.App().Get("/api", handler)
//	s.UseNotFound() // 放在路由之后
func (s *Server) UseNotFound(handler ...fiber.Handler) *Server {
	if len(handler) == 0 {
		handler = []fiber.Handler{NotFound()}
	}
	s.app.Use(handler[0])
	return s
}
