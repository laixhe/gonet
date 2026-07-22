package xfiber

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	contribJwt "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func createTestToken(t *testing.T, secret []byte) string {
	t.Helper()
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.MapClaims{
		"sub": "1234567890",
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenString
}

func TestUseJwt_DefaultErrorHandler(t *testing.T) {
	// 不传 ErrorHandler 时，默认使用 JwtErrorHandler
	secret := []byte("test-secret")
	handler := UseJwt(contribJwt.Config{
		SigningKey: contribJwt.SigningKey{
			JWTAlg: "HS256",
			Key:    secret,
		},
	})
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}

	// 使用无效 token 验证返回 401（JwtErrorHandler 生效）
	logger := newTestLogger()
	s := New(logger)
	s.App().Use(handler)
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 from JwtErrorHandler, got %d", resp.StatusCode)
	}
}

func TestUseJwt_WithErrorHandler(t *testing.T) {
	// 传入自定义 ErrorHandler 时，不应覆盖
	customCalled := false
	customHandler := func(c fiber.Ctx, err error) error {
		customCalled = true
		return c.Status(fiber.StatusForbidden).SendString("custom")
	}

	handler := UseJwt(contribJwt.Config{
		SigningKey: contribJwt.SigningKey{
			JWTAlg: "HS256",
			Key:    []byte("secret"),
		},
		ErrorHandler: customHandler,
	})
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}

	// 测试 handler 被调用时 ErrorHandler 是否生效
	logger := newTestLogger()
	s := New(logger)
	s.App().Use(handler)
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if !customCalled {
		t.Error("expected custom error handler to be called")
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

func TestUseJwt_ValidToken(t *testing.T) {
	secret := []byte("test-secret")
	token := createTestToken(t, secret)

	logger := newTestLogger()
	s := New(logger)
	s.App().Use(UseJwt(contribJwt.Config{
		SigningKey: contribJwt.SigningKey{
			JWTAlg: "HS256",
			Key:    secret,
		},
	}))
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("authorized")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUseJwt_InvalidToken(t *testing.T) {
	secret := []byte("test-secret")

	logger := newTestLogger()
	s := New(logger)
	s.App().Use(UseJwt(contribJwt.Config{
		SigningKey: contribJwt.SigningKey{
			JWTAlg: "HS256",
			Key:    secret,
		},
	}))
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	// 验证返回 JSON 格式：{"code":401,"message":"..."}
	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if body["code"] == nil {
		t.Error("expected 'code' field in error response")
	}
}

// TestJwtErrorHandler 直接调用 JwtErrorHandler，验证返回 401 JSON。
func TestJwtErrorHandler(t *testing.T) {
	logger := newTestLogger()
	s := New(logger)

	app := s.App()
	app.Get("/test", func(c fiber.Ctx) error {
		// 直接调用 JwtErrorHandler
		return JwtErrorHandler(c, jwtv5.ErrSignatureInvalid)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer some.token.here")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if body["code"] == nil {
		t.Error("expected 'code' field in response")
	}
}

func TestJwtOptionalErrorHandler(t *testing.T) {
	// JwtOptionalErrorHandler 作为 JWT 的 ErrorHandler 时，
	// JWT 校验失败不返回 401，而是放行请求继续执行后续 handler
	secret := []byte("test-secret")

	logger := newTestLogger()
	s := New(logger)
	s.App().Use(UseJwt(contribJwt.Config{
		SigningKey: contribJwt.SigningKey{
			JWTAlg: "HS256",
			Key:    secret,
		},
		ErrorHandler: JwtOptionalErrorHandler,
	}))
	s.App().Get("/test", func(c fiber.Ctx) error {
		return c.SendString("pass-through")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	resp, err := s.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 (pass-through), got %d", resp.StatusCode)
	}
}
