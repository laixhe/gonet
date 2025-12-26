package xgin

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	contribZap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/laixhe/gonet/jwt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Server struct {
	isDebug bool
	logger  *zap.Logger
	app     *gin.Engine
}

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
	s.useLog()
	return s
}

func (s *Server) App() *gin.Engine {
	return s.app
}

// UseRecover 中间件-恢复panic
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

// UseCors 中间件-跨域
func (s *Server) UseCors(config ...cors.Config) *Server {
	if len(config) == 0 {
		defaultConfig := cors.DefaultConfig()
		defaultConfig.AllowAllOrigins = true
		config = []cors.Config{defaultConfig}
	}
	s.app.Use(cors.New(config[0]))
	return s
}

// useLog 中间件-日志
func (s *Server) useLog() *Server {
	s.app.Use(contribZap.GinzapWithConfig(s.logger, &contribZap.Config{
		Context: func(ctx *gin.Context) []zapcore.Field {
			fields := make([]zapcore.Field, 0, 5)
			// log X-Request-Id
			fields = append(fields, zap.String("requestId", requestid.Get(ctx)))
			// log Content-Type
			contentType := ctx.Request.Header.Get("Content-Type")
			fields = append(fields, zap.String("contentType", contentType))
			// log Authorization
			authorization := ctx.Request.Header.Get(jwt.Authorization)
			fields = append(fields, zap.String("authorization", authorization))
			// log Body
			if ctx.Request.Method == http.MethodPost || ctx.Request.Method == http.MethodPut {
				// 如果不是文件上传类型，则读取 body
				if !strings.Contains(contentType, binding.MIMEMultipartPOSTForm) {
					// 读取 body 数据
					if body, err := ctx.GetRawData(); err == nil {
						fields = append(fields, zap.String("body", string(body)))
						// 重置 body 指针，以便后续处理
						if len(body) > 0 {
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

// Handlers404Error 处理所有未找到的路由
func (s *Server) Handlers404Error(errorFunc ...gin.HandlerFunc) *Server {
	s.app.NoRoute(func(ctx *gin.Context) {
		if len(errorFunc) == 0 {
			ctx.JSON(http.StatusNotFound, Error{
				Code:    http.StatusNotFound,
				Message: "Not Found",
			})
		} else {
			errorFunc[0](ctx)
		}
	})
	return s
}

// Listen 启动 Http 服务
func (s *Server) Listen(addr string) error {
	return s.app.Run(addr)
}
