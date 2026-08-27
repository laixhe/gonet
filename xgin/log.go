package xgin

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapField 返回常用日志字段（requestId 和请求路径），供业务层在 handler 中追加自定义字段时复用。
// 典型用法：
//
//	fields := xgin.ZapField(ctx)
//	fields = append(fields, zap.String("userId", userID))
//	logs.Logger().Info("request", fields...)
func ZapField(ctx *gin.Context) []zap.Field {
	return []zap.Field{
		zap.String("requestId", requestid.Get(ctx)),
		zap.String("path", ctx.Request.URL.Path),
	}
}
