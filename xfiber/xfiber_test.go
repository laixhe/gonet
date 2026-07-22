package xfiber

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// newTestLogger 创建一个不输出任何日志的 logger，用于常规测试。
func newTestLogger() *zap.Logger {
	return zap.NewNop()
}

// newDevelopmentLogger 创建一个输出到终端的 logger，用于需要验证日志输出的测试。
func newDevelopmentLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func TestNew(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	if s == nil {
		t.Fatal("expected non-nil Server")
	}
	if s.app == nil {
		t.Fatal("expected non-nil fiber.App")
	}
}

func TestNew_WithConfig(t *testing.T) {
	logger := newTestLogger()
	s := New(logger, fiber.Config{
		AppName: "test-app",
	})
	app := s.App()
	if app.Config().AppName != "test-app" {
		t.Errorf("expected AppName 'test-app', got '%s'", app.Config().AppName)
	}
}

func TestNew_SetsDefaultErrorHandler(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	app := s.App()
	// 验证 ErrorHandler 已设置（DefaultErrorHandler 被自动注册）
	if app.Config().ErrorHandler == nil {
		t.Error("expected ErrorHandler to be set")
	}
}

func TestApp(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	app := s.App()
	if app == nil {
		t.Fatal("expected non-nil fiber.App")
	}
}

func TestLoggerConfig(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	lc := s.LoggerConfig()
	if lc == nil {
		t.Error("expected non-nil LoggerConfig")
	}
}

func TestMaskAuth(t *testing.T) {
	tests := []struct {
		name string
		auth string
		want string
	}{
		{"empty", "", ""},
		{"bearer token", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "Bearer ***"},
		{"basic auth", "Basic dXNlcjpwYXNz", "Basic ***"},
		{"no space", "rawTokenWithoutSpace", "***"},
		{"custom scheme", "Custom abc123", "Custom ***"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskAuth(tt.auth)
			if got != tt.want {
				t.Errorf("maskAuth(%q) = %q, want %q", tt.auth, got, tt.want)
			}
		})
	}
}

func TestFullChain(t *testing.T) {
	// 测试多条中间件同时注册时互不干扰，请求正常完成
	logger := newTestLogger()
	s := New(logger).UseLog().UseCors().UseRecover().UseCompress()
	s.App().Get("/hello", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"msg": "hello"})
	})
	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/hello", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestNew_PreservesCustomErrorHandler(t *testing.T) {
	// 用户已设置 ErrorHandler 时 New 不应覆盖；用 418 (StatusTeapot) 作唯一标识验证
	logger := newTestLogger()
	customHandler := func(c fiber.Ctx, err error) error {
		return c.Status(fiber.StatusTeapot).SendString("custom")
	}
	s := New(logger, fiber.Config{
		ErrorHandler: customHandler,
	})
	app := s.App()
	app.Get("/test", func(c fiber.Ctx) error {
		return fiber.ErrBadRequest
	})
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusTeapot {
		t.Errorf("expected 418 (custom handler preserved), got %d", resp.StatusCode)
	}
}

func TestNew_DefaultErrorHandler_Integration(t *testing.T) {
	// 验证 DefaultErrorHandler 在完整请求链路中生效
	logger := newTestLogger()
	s := New(logger)
	s.App().Get("/error", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "resource not found")
	})
	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/error", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
