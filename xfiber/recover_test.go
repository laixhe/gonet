package xfiber

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func TestUseRecover(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	result := s.UseRecover()
	if result != s {
		t.Error("UseRecover should return self for chaining")
	}
	// 验证 panic 被 recover 捕获（不传 config 时使用默认配置）
	s.App().Get("/panic", func(c fiber.Ctx) error {
		panic("test panic")
	})
	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/panic", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("expected 500 after recover, got %d", resp.StatusCode)
	}
}

func TestUseRecover_EnableStackTrace(t *testing.T) {
	// 传入 recover.Config 启用 stack trace 记录
	logger := newTestLogger()
	s := New(logger).UseRecover(recover.Config{
		EnableStackTrace: true,
	})
	s.App().Get("/panic", func(c fiber.Ctx) error {
		panic("test panic with stack trace")
	})
	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/panic", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}
