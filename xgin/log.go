package xgin

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapField 日志字段
func ZapField(c *gin.Context) []zap.Field {
	return []zap.Field{
		zap.String(HeaderRequestID, GetRequestID(c)),
		zap.String("path", c.Request.URL.Path),
	}
}
