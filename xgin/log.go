package xgin

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapField 日志字段
func ZapField(ctx *gin.Context) []zap.Field {
	return []zap.Field{
		zap.String("requestId", requestid.Get(ctx)),
		zap.String("path", ctx.Request.URL.Path),
	}
}
