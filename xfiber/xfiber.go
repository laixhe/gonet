// Package xfiber 提供基于 gofiber/fiber 的 HTTP 服务封装，集成日志、限流、超时、JWT 鉴权、CORS、压缩等常用中间件。
//
// 典型用法：
//
//	logs, _ := xlog.InitZap(&xlog.Config{CallerSkip: 1})
//	app := xfiber.New(logs.Logger()).
//	    UseLog().UseLimiter(limiter.Config{Max: 100}).
//	    UseTimeout(timeout.Config{Timeout: 30 * time.Second}).
//	    UseCors().UseRecover()
//	app.Listen(":8010")
package xfiber

import (
	"strings"

	contribZap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"go.uber.org/zap"
)

// Server 封装 Fiber App，内置请求 ID 注入，通过链式调用 UseXxx 方法注册中间件。
// 使用 New 创建实例，通过 UseLog、UseLimiter、UseCors 等方法链式配置中间件后调用 Listen 启动。
type Server struct {
	logger       *zap.Logger
	loggerConfig *contribZap.LoggerConfig
	app          *fiber.App
}

// New 创建 Server 实例，自动注册默认错误处理与请求 ID 中间件。
// 通过 UseLog、UseLimiter、UseRecover 等链式方法配置中间件，最后调用 Listen 启动服务。
func New(logger *zap.Logger, config ...fiber.Config) *Server {
	if len(config) == 0 {
		config = []fiber.Config{{}}
	}
	if config[0].ErrorHandler == nil {
		config[0].ErrorHandler = DefaultErrorHandler()
	}
	// 创建 contribZap Logger 实例，用于 Fiber 内部日志输出
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
	// 替换 Fiber 全局日志为 zap logger，使 log.WithContext() 等调用使用用户配置的 logger
	log.SetLogger[*zap.Logger](loggerConfig)
	return s
}

// App 返回底层 *fiber.App，供需要直接操作 Fiber 实例的场景（如注册自定义路由组、中间件）。
func (s *Server) App() *fiber.App {
	return s.app
}

// LoggerConfig 返回配置好的 zap logger，供需要直接操作 Fiber 全局日志的场景（如自定义 log 输出）。
func (s *Server) LoggerConfig() *contribZap.LoggerConfig {
	return s.loggerConfig
}

// Listen 启动 HTTP 服务，监听指定地址（如 ":8010"）。
// 该方法会阻塞直到服务停止，等价于 fiber.App.Listen(addr)。
func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
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
