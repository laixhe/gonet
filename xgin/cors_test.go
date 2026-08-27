package xgin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zaptest"
)

func TestUseCors_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseCors()

	s.app.GET("/test-cors", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test-cors", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestUseCors_WithCustomConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://example.com"}
	s.UseCors(corsConfig)

	s.app.GET("/test-cors-custom", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test-cors-custom", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
