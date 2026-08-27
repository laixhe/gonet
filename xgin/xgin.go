// Package xgin 提供基于 gin-gonic/gin 的 HTTP 服务封装，集成日志、限流、超时、JWT 鉴权、CORS、压缩等常用中间件。
//
// 典型用法：
//
//	logs, _ := xlog.InitZap(&xlog.Config{CallerSkip: 1})
//	app := xgin.New(true, logs.Logger()).
//	    UseLog().UseLimiter(xgin.LimiterConfig{Rate: 100}).
//	    UseTimeout(30 * time.Second).
//	    UseCors().UseRecover()
//	app.Listen(":8010")
package xgin

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-contrib/requestid"
	contribZap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/laixhe/gonet/jwt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Server 封装 Gin Engine，内置请求 ID 注入，通过链式调用 UseXxx 方法注册中间件。
// 使用 New 创建实例，通过链式配置后调用 Listen 启动服务。
type Server struct {
	isDebug bool
	logger  *zap.Logger
	app     *gin.Engine
}

// New 创建 Server 实例并注册 requestId 中间件。
// isDebug 为 true 时启用 Gin 调试模式，否则使用发布模式。
// opts 透传给 gin.New，用于注册自定义中间件或修改引擎配置。
func New(isDebug bool, logger *zap.Logger, opts ...gin.OptionFunc) *Server {
	if isDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	s := &Server{
		isDebug: isDebug,
		logger:  logger,
		app:     gin.New(opts...),
	}
	s.app.Use(requestid.New())
	return s
}

// App 返回底层 *gin.Engine，供需要直接操作 Gin 实例的场景（如注册自定义路由组、中间件）。
func (s *Server) App() *gin.Engine {
	return s.app
}

// UseLog 注册请求日志中间件，记录 requestId、Content-Type、Authorization（脱敏）和 POST/PUT 请求体。
func (s *Server) UseLog() *Server {
	s.app.Use(contribZap.GinzapWithConfig(s.logger, &contribZap.Config{
		Context: func(ctx *gin.Context) []zapcore.Field {
			fields := make([]zapcore.Field, 0, 5)
			// log X-Request-Id
			fields = append(fields, zap.String("requestId", requestid.Get(ctx)))
			// log Content-Type
			contentType := ctx.Request.Header.Get("Content-Type")
			fields = append(fields, zap.String("contentType", contentType))
			// log Authorization（脱敏处理，避免泄露令牌原文）
			authorization := maskAuth(ctx.Request.Header.Get(jwt.Authorization))
			fields = append(fields, zap.String("authorization", authorization))
			// log Body
			if ctx.Request.Method == http.MethodPost || ctx.Request.Method == http.MethodPut {
				// 如果不是文件上传类型，则读取 body
				if !strings.Contains(contentType, binding.MIMEMultipartPOSTForm) {
					// 读取 body 数据
					if body, err := ctx.GetRawData(); err == nil {
						if len(body) > 0 {
							fields = append(fields, zap.String("body", string(body)))
							ctx.Set("RequestBody", body)
							// 重置 body 指针，以便后续处理
							ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
						}
					}
				}
			}
			return fields
		},
	}))
	return s
}

// Listen 启动 HTTP 服务，监听指定地址（如 ":8010"）。
// 该方法会阻塞直到服务停止，等价于 gin.Engine.Run(addr)。
func (s *Server) Listen(addr string) error {
	return s.app.Run(addr)
}

// maskAuth 对 Authorization header 值进行脱敏，避免日志泄露凭证。
// 例如 "Bearer eyJhbG..." => "Bearer ***"。
func maskAuth(auth string) string {
	if auth == "" {
		return ""
	}
	if idx := strings.IndexByte(auth, ' '); idx != -1 {
		return auth[:idx] + " ***"
	}
	return "***"
}
