package xgin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handlers404Error 注册 404 未匹配路由处理。不传 errorFunc 时默认返回 JSON 格式的 404 错误。
// 传入自定义 HandlerFunc 可覆盖默认行为（如返回静态 HTML 页面）。
func (s *Server) Handlers404Error(errorFunc ...gin.HandlerFunc) *Server {
	s.app.NoRoute(func(ctx *gin.Context) {
		if len(errorFunc) == 0 {
			ctx.JSON(http.StatusNotFound, NotFoundError())
		} else {
			errorFunc[0](ctx)
		}
	})
	return s
}
