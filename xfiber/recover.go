package xfiber

import (
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// UseRecover 注册 panic 恢复中间件。handler 中发生 panic 时自动捕获并返回 500，避免请求连接中断。
// 不传 config 时使用 recover 默认配置；传入 recover.Config 可自定义错误响应格式。
func (s *Server) UseRecover(config ...recover.Config) *Server {
	s.app.Use(recover.New(config...))
	return s
}
