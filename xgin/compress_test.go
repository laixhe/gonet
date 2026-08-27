package xgin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zaptest"
)

func TestUseCompress_WithGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseCompress()

	s.app.GET("/compress", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, strings.Repeat("hello", 200))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/compress", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Content-Encoding = %q, want gzip", w.Header().Get("Content-Encoding"))
	}
}

func TestUseCompress_NoAcceptEncoding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseCompress()

	s.app.GET("/nocompress", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nocompress", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Error("should not gzip when client doesn't accept it")
	}
}
