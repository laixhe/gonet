package xgin

import (
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

// UseTimeout 注册请求超时中间件。
// 超时后返回 408 JSON（TimeoutError 格式）。
//
// 用法：
//
//	s.UseTimeout(5 * time.Second)
func (s *Server) UseTimeout(d time.Duration) *Server {
	s.app.Use(timeout.New(
		timeout.WithTimeout(d),
		timeout.WithResponse(func(c *gin.Context) {
			c.JSON(http.StatusRequestTimeout, TimeoutError())
		}),
	))
	return s
}
