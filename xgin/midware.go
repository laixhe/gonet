package xgin

import (
	"bytes"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/laixhe/gonet/protocol/gen/config/clog"
	xginConstant "github.com/laixhe/gonet/xgin/constant"
	"github.com/laixhe/gonet/xjwt"
	"github.com/laixhe/gonet/xlog"
)

// 中间件

// Cors 跨域
func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},                                                         // 允许所有来源（包括子域和端口），生产环境中应替换为域名
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},                   // 允许的 HTTP 方法列表，如 GET、POST、PUT 等，默认为 ["*"]（全部允许）
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},                   // 允许的 HTTP 头部列表，默认为 ["*"]（全部允许），可以自定义字段
		ExposeHeaders:    []string{"Origin", "Content-Type", "Content-Length", "Authorization"}, // 允许浏览器（客户端）可以解析的头部
		AllowCredentials: true,                                                                  // 是否允许浏览器发送 Cookie 默认为 false
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour, // 预检请求（OPTIONS）的缓存时间（秒）。默认为5分钟
	})
}

// Logger 日志
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if xlog.GetLevel() == clog.LevelType_debug.String() {
			contentType := c.Request.Header.Get("Content-Type")
			var body []byte
			var err error
			if c.Request.Method != http.MethodGet {
				// 如果不是文件上传类型，则读取body
				if !strings.Contains(contentType, "multipart/form-data") {
					// 读取body数据
					body, err = c.GetRawData()
					if err != nil {
						xlog.Error("get body error", zap.Error(err))
					} else {
						// 重置 body 指针，以便后续处理
						if len(body) > 0 {
							c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
						}
					}
				}
			}
			xlog.Debug("gin",
				zap.String(xginConstant.HeaderRequestID, requestid.Get(c)),
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("query", c.Request.URL.RawQuery),
				zap.String("ip", c.ClientIP()),
				zap.String("agent", c.Request.UserAgent()),
				zap.String("authorization", c.Request.Header.Get(xjwt.Authorization)),
				zap.String("content-type", contentType),
				zap.ByteString("body", body),
			)
		}
		c.Next()
	}
}

// Recovery 出现 panic 恢复正常
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		xlog.Error("panic",
			zap.String(xginConstant.HeaderRequestID, requestid.Get(c)),
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("agent", c.Request.UserAgent()),
			zap.String("authorization", c.Request.Header.Get(xjwt.Authorization)),
			zap.Any("error", err),
			zap.String("stack", string(debug.Stack())),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
