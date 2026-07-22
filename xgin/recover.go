package xgin

import (
	contribZap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
)

// UseRecover 注册 panic 恢复中间件。传入自定义 errorFunc 可覆盖默认的 ErrorRecoveryFunc。
// 如果 logger 不为 nil，则使用 zap 输出 panic 日志；否则使用 Gin 默认方式。
func (s *Server) UseRecover(errorFunc ...gin.RecoveryFunc) *Server {
	if s.logger != nil {
		if len(errorFunc) > 0 {
			s.app.Use(contribZap.CustomRecoveryWithZap(s.logger, true, errorFunc[0]))
		} else {
			s.app.Use(contribZap.CustomRecoveryWithZap(s.logger, true, ErrorRecoveryFunc))
		}
	} else {
		if len(errorFunc) > 0 {
			s.app.Use(gin.CustomRecovery(errorFunc[0]))
		} else {
			s.app.Use(gin.CustomRecovery(ErrorRecoveryFunc))
		}
	}
	return s
}
