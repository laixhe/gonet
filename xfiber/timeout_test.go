package xfiber

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/timeout"
)

func TestUseTimeout_WithinTimeout(t *testing.T) {
	logger := newTestLogger()
	s := New(logger).UseTimeout(timeout.Config{
		Timeout: 1 * time.Second,
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

// TestUseTimeout_Exceeded 请求处理超过超时时间，验证返回 408 + TimeoutError JSON。
func TestUseTimeout_Exceeded(t *testing.T) {
	logger := newTestLogger()
	s := New(logger).UseTimeout(timeout.Config{
		Timeout: 50 * time.Millisecond,
	})

	s.App().Get("/slow", func(c fiber.Ctx) error {
		time.Sleep(200 * time.Millisecond)
		return c.SendString("too late")
	})

	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/slow", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusRequestTimeout {
		t.Errorf("expected 408, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if body["message"] != "Request Timeout" {
		t.Errorf("expected 'Request Timeout' message, got '%v'", body["message"])
	}
}

func TestUseTimeout_ReturnsSelfForChaining(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	result := s.UseTimeout(timeout.Config{
		Timeout: 5 * time.Second,
	})
	if result != s {
		t.Error("UseTimeout should return self for chaining")
	}
}

func TestUseTimeout_NoConfigTimeout(t *testing.T) {
	// 未设置 Timeout 则无超时限制，请求正常完成
	logger := newTestLogger()
	s := New(logger).UseTimeout()

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

func TestUseTimeout_CustomOnTimeout(t *testing.T) {
	logger := newTestLogger()
	s := New(logger).UseTimeout(timeout.Config{
		Timeout: 10 * time.Millisecond,
		OnTimeout: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusRequestTimeout).SendString("custom timeout")
		},
	})

	s.App().Get("/slow", func(c fiber.Ctx) error {
		time.Sleep(100 * time.Millisecond)
		return c.SendString("too late")
	})

	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/slow", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusRequestTimeout {
		t.Errorf("expected 408, got %d", resp.StatusCode)
	}
}

func TestUseTimeout_ChainIntegration(t *testing.T) {
	logger := newTestLogger()
	s := New(logger).UseLog().UseCors().UseRecover().UseTimeout(timeout.Config{
		Timeout: 1 * time.Second,
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

func TestTimeoutError_Default(t *testing.T) {
	err := TimeoutError()
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Code != fiber.StatusRequestTimeout {
		t.Errorf("expected 408, got %d", err.Code)
	}
	if err.Message != "Request Timeout" {
		t.Errorf("expected 'Request Timeout', got '%s'", err.Message)
	}
}

func TestTimeoutError_CustomMessage(t *testing.T) {
	err := TimeoutError("处理超时")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Code != fiber.StatusRequestTimeout {
		t.Errorf("expected 408, got %d", err.Code)
	}
	if err.Message != "处理超时" {
		t.Errorf("expected '处理超时', got '%s'", err.Message)
	}
}
