package xfiber

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/timeout"
)

// Timeout 返回请求超时中间件，封装 fiber 官方 timeout 中间件。
// 不传 config 时使用默认配置（OnTimeout 返回 TimeoutError JSON 响应）。
// 传入 timeout.Config 可自定义超时时间、OnTimeout 回调、跳过路径等。
func Timeout(config ...timeout.Config) fiber.Handler {
	if len(config) == 0 {
		config = []timeout.Config{{}}
	}
	if config[0].OnTimeout == nil {
		config[0].OnTimeout = func(c fiber.Ctx) error {
			return c.Status(fiber.StatusRequestTimeout).JSON(TimeoutError())
		}
	}
	return timeout.New(func(c fiber.Ctx) error {
		return c.Next()
	}, config[0])
}

// UseTimeout 注册请求超时中间件。
//
// 用法：
//
//	s.UseTimeout(timeout.Config{Timeout: 5 * time.Second})
//
// 不传 config 时不设超时时间（需自行指定 Timeout）；超时后返回 408 JSON。
func (s *Server) UseTimeout(config ...timeout.Config) *Server {
	s.app.Use(Timeout(config...))
	return s
}
