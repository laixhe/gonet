package xfiber

import (
	"github.com/gofiber/fiber/v3/middleware/compress"
)

// UseCompress 注册响应压缩中间件。不传 config 时使用 compress 默认配置（gzip 压缩）。
// 传入 compress.Config 可自定义压缩级别（Level）、跳过条件（Next）等策略。
func (s *Server) UseCompress(config ...compress.Config) *Server {
	s.app.Use(compress.New(config...))
	return s
}
