package xfiber

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
)

func TestUseCompress(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	result := s.UseCompress()
	if result != s {
		t.Error("UseCompress should return self for chaining")
	}
	s.App().Get("/compress", func(c fiber.Ctx) error {
		return c.SendString(strings.Repeat("hello", 200)) // 足够长的响应体触发压缩
	})
	req := httptest.NewRequest(fiber.MethodGet, "/compress", nil)
	req.Header.Set(fiber.HeaderAcceptEncoding, "gzip")
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get(fiber.HeaderContentEncoding) != "gzip" {
		t.Errorf("expected Content-Encoding 'gzip', got '%s'", resp.Header.Get(fiber.HeaderContentEncoding))
	}
}

func TestUseCompress_WithConfig(t *testing.T) {
	logger := newTestLogger()
	s := New(logger).UseCompress(compress.Config{
		Level: compress.LevelBestSpeed,
	})
	s.App().Get("/compress", func(c fiber.Ctx) error {
		return c.SendString(strings.Repeat("world", 200))
	})
	req := httptest.NewRequest(fiber.MethodGet, "/compress", nil)
	req.Header.Set(fiber.HeaderAcceptEncoding, "gzip")
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get(fiber.HeaderContentEncoding) != "gzip" {
		t.Errorf("expected Content-Encoding 'gzip', got '%s'", resp.Header.Get(fiber.HeaderContentEncoding))
	}
}

// TestUseCompress_NoAcceptEncoding 客户端未声明 Accept-Encoding 时，响应不被压缩。
func TestUseCompress_NoAcceptEncoding(t *testing.T) {
	logger := newTestLogger()
	s := New(logger).UseCompress()
	s.App().Get("/nocompress", func(c fiber.Ctx) error {
		return c.SendString("hello")
	})
	req := httptest.NewRequest(fiber.MethodGet, "/nocompress", nil)
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get(fiber.HeaderContentEncoding) == "gzip" {
		t.Error("expected no Content-Encoding when client does not accept gzip")
	}
}
