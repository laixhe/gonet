package xfiber

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// Limiter 返回全局限流中间件，基于客户端 IP 限流，封装 fiber 官方 limiter 中间件。
// 默认使用 SlidingWindow 滑动窗口策略，KeyGenerator 为 c.IP()，超限返回 RateLimitError JSON。
// 传入 limiter.Config 可自定义阈值（Max）、窗口时间（Expiration）、跳过条件（Next）等。
func Limiter(config ...limiter.Config) fiber.Handler {
	if len(config) == 0 {
		config = []limiter.Config{{}}
	}
	if config[0].LimiterMiddleware == nil {
		config[0].LimiterMiddleware = limiter.SlidingWindow{}
	}
	if config[0].LimitReached == nil {
		config[0].LimitReached = func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(RateLimitError())
		}
	}
	return limiter.New(config[0])
}

// UseLimiter 注册全局限流中间件。
// 不传 config 时使用默认配置（滑动窗口，Max=5，窗口 1 分钟）。
// 传入 limiter.Config 可自定义阈值、窗口时间、限流策略等。
func (s *Server) UseLimiter(config ...limiter.Config) *Server {
	s.app.Use(Limiter(config...))
	return s
}
