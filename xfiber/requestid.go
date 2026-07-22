package xfiber

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// RequestIdLogKey 请求 ID 日志 Key
const RequestIdLogKey = "requestId"

// useRequestId 注册请求 ID 中间件，自动注入 requestId 到 context 日志 key 中。
// New() 已默认调用，一般无需手动调用；重复调用不产生副作用。
func (s *Server) useRequestId() *Server {
	s.app.Use(requestid.New())
	s.app.Use(func(ctx fiber.Ctx) error {
		newCtx := context.WithValue(ctx.Context(), RequestIdLogKey, ctx.GetRespHeader(fiber.HeaderXRequestID))
		ctx.SetContext(newCtx)
		return ctx.Next()
	})
	return s
}
