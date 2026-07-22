package xfiber

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func TestUseCors(t *testing.T) {
	// 使用默认 CORS 配置（无 Origin 限制），OPTIONS 预检应返回 200 或 204
	logger := newTestLogger()
	s := New(logger)
	result := s.UseCors()
	if result != s {
		t.Error("UseCors should return self for chaining")
	}
	// CORS 预检请求应返回 204 且包含 CORS 头
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	req := httptest.NewRequest(fiber.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusNoContent {
		t.Errorf("expected 200/204 for OPTIONS, got %d", resp.StatusCode)
	}
	t.Logf("CORS response status: %d", resp.StatusCode)
}

func TestUseCors_WithConfig(t *testing.T) {
	// 限定特定 Origin，验证 OPTIONS 和正常 GET 均返回正确的 CORS 头
	logger := newTestLogger()
	s := New(logger).UseCors(cors.Config{
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET", "POST"},
	})
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// OPTIONS 预检
	req := httptest.NewRequest(fiber.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin != "https://example.com" {
		t.Errorf("expected Allow-Origin 'https://example.com', got '%s'", allowOrigin)
	}

	// 正常 GET 请求也应带 CORS 头
	req2 := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	req2.Header.Set("Origin", "https://example.com")
	resp2, err := s.App().Test(req2)
	if err != nil {
		t.Fatal(err)
	}
	if resp2.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp2.StatusCode)
	}
	if resp2.Header.Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Error("expected CORS header on normal GET request")
	}
}
