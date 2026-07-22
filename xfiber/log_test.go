package xfiber

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestUseLog(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	result := s.UseLog()
	if result != s {
		t.Error("UseLog should return self for chaining")
	}
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

func TestUseLog_SkipsMultipartBody(t *testing.T) {
	// 仅设 Content-Type 头（无实际 body），验证 SkipBody 决策能正常工作
	logger := newTestLogger()
	s := New(logger).UseLog()
	s.App().Post("/upload", func(c fiber.Ctx) error {
		return c.SendString("uploaded")
	})
	req := httptest.NewRequest(fiber.MethodPost, "/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUseLog_FieldsFunc(t *testing.T) {
	// 发起带 Content-Type 和 Authorization 的请求，触发 FieldsFunc 闭包
	// 使用 development logger 确保中间件内部回调被完整执行
	logger := newDevelopmentLogger()
	s := New(logger).UseLog()
	s.App().Post("/api", func(c fiber.Ctx) error {
		return c.SendString("created")
	})

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(`{"key":"value"}`))
	req.Header.Set(fiber.HeaderContentType, "application/json")
	req.Header.Set(fiber.HeaderAuthorization, "Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjMifQ.xxx")
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUseLog_SkipBody_Multipart(t *testing.T) {
	// 发送真实 multipart body 请求，验证 body 被跳过且不影响正常处理
	logger := newTestLogger()
	s := New(logger).UseLog()
	s.App().Post("/upload", func(c fiber.Ctx) error {
		return c.SendString("uploaded")
	})

	body := strings.NewReader("--boundary\r\nContent-Disposition: form-data; name=\"file\"\r\n\r\ncontent\r\n--boundary--\r\n")
	req := httptest.NewRequest(fiber.MethodPost, "/upload", body)
	req.Header.Set(fiber.HeaderContentType, "multipart/form-data; boundary=boundary")
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
