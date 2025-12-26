package xfiber

import (
	"context"

	contribZap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"go.uber.org/zap"
)

// RequestIdLogKey 请求 ID 日志 Key
const RequestIdLogKey = "requestId"

type Server struct {
	logger       *zap.Logger
	loggerConfig *contribZap.LoggerConfig
	app          *fiber.App
}

func New(logger *zap.Logger, config ...fiber.Config) *Server {
	if len(config) == 0 {
		config = []fiber.Config{{}}
	}
	if config[0].ErrorHandler == nil {
		config[0].ErrorHandler = DefaultErrorHandler()
	}
	// 默认日志
	loggerConfig := contribZap.NewLogger(contribZap.LoggerConfig{
		ExtraKeys: []string{RequestIdLogKey},
		SetLogger: logger,
	})
	s := &Server{
		logger:       logger,
		loggerConfig: loggerConfig,
		app:          fiber.New(config...),
	}
	s.useRequestId()
	s.useLog()
	// 替换默认日志
	log.SetLogger[*zap.Logger](loggerConfig)
	return s
}

func (s *Server) App() *fiber.App {
	return s.app
}

func (s *Server) LoggerConfig() *contribZap.LoggerConfig {
	return s.loggerConfig
}

// UseRecover 中间件-恢复panic
func (s *Server) UseRecover(config ...recover.Config) *Server {
	s.app.Use(recover.New(config...))
	return s
}

// UseCors 中间件-跨域
func (s *Server) UseCors(config ...cors.Config) *Server {
	s.app.Use(cors.New(config...))
	return s
}

// useRequestId 中间件-请求ID
func (s *Server) useRequestId() *Server {
	s.app.Use(requestid.New())
	s.app.Use(func(ctx fiber.Ctx) error {
		newCtx := context.WithValue(ctx.Context(), RequestIdLogKey, ctx.GetRespHeader(fiber.HeaderXRequestID))
		ctx.SetContext(newCtx)
		return ctx.Next()
	})
	return s
}

// useLog 中间件-日志
func (s *Server) useLog() *Server {
	config := contribZap.Config{
		Logger: s.logger,
		Fields: []string{"ip", "latency", "status", RequestIdLogKey, "method", "url", "body"},
		FieldsFunc: func(ctx fiber.Ctx) []zap.Field {
			fields := []zap.Field{
				zap.String("contentType", ctx.Get(fiber.HeaderContentType)),
				zap.String("authorization", ctx.Get(fiber.HeaderAuthorization)),
			}
			return fields
		},
	}
	s.app.Use(contribZap.New(config))
	return s
}

// Listen 启动 Http 服务
func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}
