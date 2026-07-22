package xfiber

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// TestUseNotFound_Default 已匹配路由正常返回，未匹配路由返回 {"code":404,"message":"Not Found"}。
func TestUseNotFound_Default(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	s.App().Get("/api", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	s.UseNotFound()

	// 已匹配路由正常响应
	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/api", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 for matched route, got %d", resp.StatusCode)
	}

	// 未匹配路由返回 404 JSON
	resp2, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/nonexistent", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp2.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404 for unmatched route, got %d", resp2.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp2.Body).Decode(&body)
	if body["message"] != "Not Found" {
		t.Errorf("expected 'Not Found' message, got '%v'", body["message"])
	}
	if body["code"] == nil {
		t.Error("expected 'code' field in response")
	}
}

func TestUseNotFound_CustomHandler(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	s.App().Get("/api", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	s.UseNotFound(func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).SendString("custom 404")
	})

	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/nonexistent", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUseNotFound_ReturnsSelf(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	result := s.UseNotFound()
	if result != s {
		t.Error("UseNotFound should return self for chaining")
	}
}

func TestNotFoundError_Default(t *testing.T) {
	err := NotFoundError()
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", err.Code)
	}
	if err.Message != "Not Found" {
		t.Errorf("expected 'Not Found', got '%s'", err.Message)
	}
}

func TestNotFoundError_CustomMessage(t *testing.T) {
	err := NotFoundError("资源不存在")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", err.Code)
	}
	if err.Message != "资源不存在" {
		t.Errorf("expected '资源不存在', got '%s'", err.Message)
	}
}
