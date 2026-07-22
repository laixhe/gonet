package xgin

import (
	"github.com/gin-contrib/cors"
)

// UseCors 注册跨域中间件。不传 config 时默认允许所有来源（AllowAllOrigins = true）。
// 传入 cors.Config 可自定义允许的方法、请求头等策略。
func (s *Server) UseCors(config ...cors.Config) *Server {
	if len(config) == 0 {
		defaultConfig := cors.DefaultConfig()
		defaultConfig.AllowAllOrigins = true
		config = []cors.Config{defaultConfig}
	}
	s.app.Use(cors.New(config[0]))
	return s
}
