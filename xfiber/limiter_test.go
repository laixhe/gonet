package xfiber

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func TestUseLimiter_WithinLimit(t *testing.T) {
	logger := newTestLogger()
	s := New(logger).UseLimiter(limiter.Config{
		Max:        3,
		Expiration: 1 * time.Second,
	})
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 前 3 次请求应全部通过
	for i := range 3 {
		resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, resp.StatusCode)
		}
	}
}

// TestUseLimiter_Exceeded 超过 Max 限制后返回 429。
func TestUseLimiter_Exceeded(t *testing.T) {
	logger := newTestLogger()
	s := New(logger).UseLimiter(limiter.Config{
		Max:        2,
		Expiration: 1 * time.Second,
	})
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 快速连续发送 3 次请求，第 3 次应被限流
	var lastStatus int
	for i := range 3 {
		resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
		if err != nil {
			t.Fatal(err)
		}
		lastStatus = resp.StatusCode
		if i < 2 && resp.StatusCode != fiber.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, resp.StatusCode)
		}
	}

	if lastStatus != fiber.StatusTooManyRequests {
		t.Errorf("expected 429 on exceeded request, got %d", lastStatus)
	}
}

func TestUseLimiter_429JSONFormat(t *testing.T) {
	logger := newTestLogger()
	s := New(logger).UseLimiter(limiter.Config{
		Max:        1,
		Expiration: 1 * time.Second,
	})
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 第 1 次通过
	s.App().Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))

	// 第 2 次应返回 429 JSON
	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if body["message"] != "Too Many Requests" {
		t.Errorf("expected 'Too Many Requests', got '%v'", body["message"])
	}
	if body["code"] == nil {
		t.Error("expected 'code' field in response")
	}
}

func TestUseLimiter_ReturnsSelfForChaining(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	result := s.UseLimiter(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Second,
	})
	if result != s {
		t.Error("UseLimiter should return self for chaining")
	}
}

func TestUseLimiter_ChainIntegration(t *testing.T) {
	logger := newTestLogger()
	s := New(logger).UseLog().UseLimiter(limiter.Config{
		Max:        3,
		Expiration: 1 * time.Second,
	})

	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUseLimiter_NoConfig(t *testing.T) {
	// 不传 config 时使用默认配置（Max=5, Expiration=1min）
	logger := newTestLogger()
	s := New(logger).UseLimiter()

	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRateLimitError_Default(t *testing.T) {
	err := RateLimitError()
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Code != fiber.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", err.Code)
	}
	if err.Message != "Too Many Requests" {
		t.Errorf("expected 'Too Many Requests', got '%s'", err.Message)
	}
}

func TestRateLimitError_CustomMessage(t *testing.T) {
	err := RateLimitError("请求过于频繁")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Code != fiber.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", err.Code)
	}
	if err.Message != "请求过于频繁" {
		t.Errorf("expected '请求过于频繁', got '%s'", err.Message)
	}
}
