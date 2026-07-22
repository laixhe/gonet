package xgin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zaptest"
)

func TestUseTimeout_WithinTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger).UseTimeout(1 * time.Second)

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

func TestUseTimeout_Exceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger).UseTimeout(50 * time.Millisecond)

	s.app.GET("/slow", func(ctx *gin.Context) {
		time.Sleep(200 * time.Millisecond)
		ctx.String(http.StatusOK, "too late")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusRequestTimeout {
		t.Errorf("status = %d, want %d", w.Code, http.StatusRequestTimeout)
	}
}

func TestUseTimeout_ZeroDuration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger).UseTimeout(0)

	s.app.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusRequestTimeout {
		t.Errorf("status = %d, want %d", w.Code, http.StatusRequestTimeout)
	}
}
