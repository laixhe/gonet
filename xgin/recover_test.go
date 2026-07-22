package xgin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zaptest"
)

func TestUseRecover_WithLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseRecover()

	s.app.GET("/panic", func(ctx *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestUseRecover_CustomFunc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseRecover(func(ctx *gin.Context, err any) {
		ctx.String(http.StatusInternalServerError, "custom recovery")
	})

	s.app.GET("/panic", func(ctx *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if w.Body.String() != "custom recovery" {
		t.Errorf("body = %q, want %q", w.Body.String(), "custom recovery")
	}
}

func TestUseRecover_WithoutLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := New(true, nil)
	s.UseRecover()

	s.app.GET("/panic", func(ctx *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestUseRecover_WithoutLogger_CustomFunc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := New(true, nil)
	s.UseRecover(func(ctx *gin.Context, err any) {
		ctx.String(http.StatusInternalServerError, "no-logger-custom")
	})

	s.app.GET("/panic-nologger", func(ctx *gin.Context) {
		panic("test panic no logger")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic-nologger", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if w.Body.String() != "no-logger-custom" {
		t.Errorf("body = %q, want %q", w.Body.String(), "no-logger-custom")
	}
}
