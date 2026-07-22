package xgin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zaptest"
)

func TestUseLimiter_WithinLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger).UseLimiter(LimiterConfig{Rate: 100, Burst: 200})

	s.app.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	for i := range 5 {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		s.app.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: status = %d, want %d", i, w.Code, http.StatusOK)
		}
	}
}

func TestUseLimiter_Exceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	// Rate=1 means 1 token/sec; since Burst=1 and we send 3 rapid requests, 3rd should fail
	s := New(true, logger).UseLimiter(LimiterConfig{Rate: 1, Burst: 1})

	s.app.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	var lastStatus int
	for range 3 {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		s.app.ServeHTTP(w, req)
		lastStatus = w.Code
	}

	if lastStatus != http.StatusTooManyRequests {
		t.Errorf("last status = %d, want %d", lastStatus, http.StatusTooManyRequests)
	}
}

func TestUseLimiter_429JSONFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger).UseLimiter(LimiterConfig{Rate: 1, Burst: 1})

	s.app.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	// First request consumes the token
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	s.app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("first request: status = %d, want %d", w.Code, http.StatusOK)
	}

	// Second should be rate limited
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	s.app.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w2.Code, http.StatusTooManyRequests)
	}
	if w2.Body.String() == "" {
		t.Error("expected JSON body")
	}
}

func TestUseLimiter_NoConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger).UseLimiter()

	s.app.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestUseLimiter_ZeroConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger).UseLimiter(LimiterConfig{Rate: 0, Burst: 0})

	s.app.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}
