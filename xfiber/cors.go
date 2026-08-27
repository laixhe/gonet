package xfiber

import (
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// UseCors 注册跨域中间件。不传 config 时使用 CORS 默认配置。
// 传入 cors.Config 可自定义允许的来源、方法、请求头等策略。
func (s *Server) UseCors(config ...cors.Config) *Server {
	s.app.Use(cors.New(config...))
	return s
}
