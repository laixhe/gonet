package xfiber

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestRequestId(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)
	// useRequestId 在 New 中自动调用，验证 requestId 头存在
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	if err != nil {
		t.Fatal(err)
	}
	reqID := resp.Header.Get(fiber.HeaderXRequestID)
	if reqID == "" {
		t.Error("expected X-Request-Id header to be set")
	}
}

func TestRequestId_RepeatedCall(t *testing.T) {
	// useRequestId 重复调用不产生副作用
	logger := newTestLogger()
	s := New(logger)
	// 手动再调用一次
	s.useRequestId()
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	resp, err := s.App().Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	if err != nil {
		t.Fatal(err)
	}
	reqID := resp.Header.Get(fiber.HeaderXRequestID)
	if reqID == "" {
		t.Error("expected X-Request-Id header after repeated useRequestId")
	}
}
