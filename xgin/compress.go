package xgin

import (
	"github.com/gin-contrib/gzip"
)

// UseCompress 注册 GZIP 响应压缩中间件。
// 不传 config 时使用默认配置（gzip.DefaultCompression）。
// 传入 gzip.Option 可自定义压缩级别、排除路径等策略。
func (s *Server) UseCompress(opts ...gzip.Option) *Server {
	s.app.Use(gzip.Gzip(gzip.DefaultCompression, opts...))
	return s
}
