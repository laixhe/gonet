package xgin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zaptest"
)

func TestHandlers404Error_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.Handlers404Error()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var resp Error
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != http.StatusNotFound {
		t.Errorf("resp.Code = %d, want %d", resp.Code, http.StatusNotFound)
	}
	if resp.Message != "Not Found" {
		t.Errorf("resp.Message = %q, want %q", resp.Message, "Not Found")
	}
}

func TestHandlers404Error_Custom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.Handlers404Error(func(ctx *gin.Context) {
		ctx.String(http.StatusNotFound, "custom 404")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	if w.Body.String() != "custom 404" {
		t.Errorf("body = %q, want %q", w.Body.String(), "custom 404")
	}
}
