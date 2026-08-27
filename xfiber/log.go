package xfiber

import (
	contribZap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// UseLog 注册请求日志中间件，自动记录每个请求的 ip、耗时、状态码、requestId、method、url、body。
// Authorization header 自动脱敏（"Bearer xxx" → "Bearer ***"），multipart 类型跳过 body 记录。
func (s *Server) UseLog() *Server {
	config := contribZap.Config{
		Logger: s.logger,
		Fields: []string{"ip", "latency", "status", RequestIdLogKey, "method", "url", "body"},
		FieldsFunc: func(ctx fiber.Ctx) []zap.Field {
			fields := []zap.Field{
				zap.String("contentType", ctx.Get(fiber.HeaderContentType)),
				zap.String("authorization", maskAuth(ctx.Get(fiber.HeaderAuthorization))),
			}
			return fields
		},
		SkipBody: func(ctx fiber.Ctx) bool {
			return ctx.IsMultipart()
		},
	}
	s.app.Use(contribZap.New(config))
	return s
}
