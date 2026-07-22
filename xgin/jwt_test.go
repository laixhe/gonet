package xgin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/laixhe/gonet/jwt"
)

// testJwtConfig 返回测试用的 JWT 配置。
func testJwtConfig() *jwt.Config {
	return &jwt.Config{
		SecretKey:     "test-secret-key-for-testing",
		ExpireTime:    3600,
		SigningMethod: jwt.SigningMethodHS256,
	}
}

// generateTestToken 生成一个有效的测试 JWT token。
func generateTestToken(t *testing.T) string {
	t.Helper()
	config := testJwtConfig()
	claims := &jwt.CustomClaims{
		Uid:              12345,
		RegisteredClaims: jwtv5.RegisteredClaims{},
	}
	token, err := jwt.GenToken(config, claims)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}
	return token
}

func TestUseJwt_NoAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()

	router := gin.New()
	router.Use(UseJwt(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		_, exists := ctx.Get(jwt.AuthorizationClaims)
		if exists {
			ctx.String(http.StatusOK, "authorized")
		} else {
			ctx.String(http.StatusOK, "no claims")
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "no claims" {
		t.Errorf("body = %q, want %q", w.Body.String(), "no claims")
	}
}

func TestUseJwt_EmptyAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()

	router := gin.New()
	router.Use(UseJwt(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		_, exists := ctx.Get(jwt.AuthorizationClaims)
		if exists {
			ctx.String(http.StatusOK, "authorized")
		} else {
			ctx.String(http.StatusOK, "no claims")
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(jwt.Authorization, "")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "no claims" {
		t.Errorf("body = %q, want %q", w.Body.String(), "no claims")
	}
}

func TestUseJwt_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()

	router := gin.New()
	router.Use(UseJwt(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		_, exists := ctx.Get(jwt.AuthorizationClaims)
		if exists {
			ctx.String(http.StatusOK, "authorized")
		} else {
			ctx.String(http.StatusOK, "no claims")
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(jwt.Authorization, jwt.Bearer+"invalid.token.here")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "no claims" {
		t.Errorf("body = %q, want %q", w.Body.String(), "no claims")
	}
}

func TestUseJwt_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()
	token := generateTestToken(t)

	router := gin.New()
	router.Use(UseJwt(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		val, exists := ctx.Get(jwt.AuthorizationClaims)
		if exists {
			if claims, ok := val.(*jwtv5.Token); ok {
				if customClaims, ok := claims.Claims.(*jwt.CustomClaims); ok {
					if customClaims.Uid == 12345 {
						ctx.String(http.StatusOK, "authorized")
						return
					}
				}
			}
		}
		ctx.String(http.StatusOK, "no claims")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(jwt.Authorization, jwt.Bearer+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "authorized" {
		t.Errorf("body = %q, want %q", w.Body.String(), "authorized")
	}
}

func TestUseJwt_AuthorizationWithoutBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()

	router := gin.New()
	router.Use(UseJwt(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		_, exists := ctx.Get(jwt.AuthorizationClaims)
		if exists {
			ctx.String(http.StatusOK, "authorized")
		} else {
			ctx.String(http.StatusOK, "no claims")
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(jwt.Authorization, "Token some-value")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "no claims" {
		t.Errorf("body = %q, want %q", w.Body.String(), "no claims")
	}
}

func TestUseJwt_ChainedToNextHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()
	token := generateTestToken(t)

	var calls []string
	router := gin.New()
	router.Use(UseJwt(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		calls = append(calls, "handler1")
	}, func(ctx *gin.Context) {
		calls = append(calls, "handler2")
		ctx.String(http.StatusOK, "done")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(jwt.Authorization, jwt.Bearer+token)
	router.ServeHTTP(w, req)

	if len(calls) != 2 {
		t.Fatalf("expected 2 handler calls, got %d", len(calls))
	}
	if calls[0] != "handler1" {
		t.Errorf("calls[0] = %q, want handler1", calls[0])
	}
	if calls[1] != "handler2" {
		t.Errorf("calls[1] = %q, want handler2", calls[1])
	}
}

func TestJwtRequired_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()

	router := gin.New()
	router.Use(JwtRequired(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJwtRequired_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()

	router := gin.New()
	router.Use(JwtRequired(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(jwt.Authorization, jwt.Bearer+"invalid.token.here")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJwtRequired_WithoutBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()

	router := gin.New()
	router.Use(JwtRequired(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(jwt.Authorization, "Token some-value")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJwtRequired_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()
	token := generateTestToken(t)

	router := gin.New()
	router.Use(JwtRequired(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "authorized")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(jwt.Authorization, jwt.Bearer+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestJwtRequired_EmptyAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()

	router := gin.New()
	router.Use(JwtRequired(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(jwt.Authorization, "")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUseJwt_ExpiredToken(t *testing.T) {
	// 使用签名无效的 token（模拟过期/伪造），验证可选模式下不阻断请求
	gin.SetMode(gin.TestMode)
	config := testJwtConfig()

	router := gin.New()
	router.Use(UseJwt(config, &jwt.CustomClaims{}))
	router.GET("/test", func(ctx *gin.Context) {
		_, exists := ctx.Get(jwt.AuthorizationClaims)
		if exists {
			ctx.String(http.StatusOK, "authorized")
		} else {
			ctx.String(http.StatusOK, "no claims")
		}
	})

	// Test with expired token - use a token that's structurally valid but has a bad signature
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(jwt.Authorization, jwt.Bearer+"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEyMzQ1fQ.badSignature")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "no claims" {
		t.Errorf("body = %q, want %q", w.Body.String(), "no claims")
	}
}
