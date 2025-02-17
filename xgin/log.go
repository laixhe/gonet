package xgin

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/laixhe/gonet/xlog"
)

// ZapField 日志字段
func ZapField(c *gin.Context) []zap.Field {
	return []zap.Field{
		zap.String(HeaderRequestID, GetRequestID(c)),
		zap.String("path", c.Request.URL.Path),
	}
}

// LogDebug 调试
func LogDebug(c *gin.Context, msg string) {
	xlog.GetLogger().Debug(msg, ZapField(c)...)
}

// LogDebugf 调试
func LogDebugf(c *gin.Context, template string, args ...interface{}) {
	xlog.GetLogger().Debug(fmt.Sprintf(template, args...), ZapField(c)...)
}

// LogInfo 信息
func LogInfo(c *gin.Context, msg string) {
	xlog.GetLogger().Info(msg, ZapField(c)...)
}

// LogInfof 信息
func LogInfof(c *gin.Context, template string, args ...interface{}) {
	xlog.GetLogger().Info(fmt.Sprintf(template, args...), ZapField(c)...)
}

// LogWarn 警告
func LogWarn(c *gin.Context, msg string) {
	xlog.GetLogger().Warn(msg, ZapField(c)...)
}

// LogWarnf 警告
func LogWarnf(c *gin.Context, template string, args ...interface{}) {
	xlog.GetLogger().Warn(fmt.Sprintf(template, args...), ZapField(c)...)
}

// LogError 错误
func LogError(c *gin.Context, msg string) {
	xlog.GetLogger().Error(msg, ZapField(c)...)
}

// LogErrorf 错误
func LogErrorf(c *gin.Context, template string, args ...interface{}) {
	xlog.GetLogger().Error(fmt.Sprintf(template, args...), ZapField(c)...)
}
