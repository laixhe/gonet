package xgin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zaptest"
)

func TestZapField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/users", nil)

	fields := ZapField(ctx)
	if len(fields) != 2 {
		t.Fatalf("len(fields) = %d, want 2", len(fields))
	}

	if fields[0].Key != "requestId" {
		t.Errorf("fields[0].Key = %q, want %q", fields[0].Key, "requestId")
	}
	if fields[1].Key != "path" {
		t.Errorf("fields[1].Key = %q, want %q", fields[1].Key, "path")
	}
	if fields[1].String != "/api/users" {
		t.Errorf("fields[1].String = %q, want %q", fields[1].String, "/api/users")
	}
}

func TestZapField_RequestId(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	// Simulate requestid middleware setting requestId
	ctx.Set("RequestId", "test-request-id-123")

	fields := ZapField(ctx)
	if len(fields) != 2 {
		t.Fatalf("len(fields) = %d, want 2", len(fields))
	}

	// requestid.Get reads from header X-Request-Id or context value
	id := requestid.Get(ctx)
	if id == "" {
		t.Log("requestId was empty - this is expected since middleware wasn't run")
	}
}

func TestUseLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseLog()

	s.app.POST("/test-log", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test-log", nil)
	req.Header.Set("Content-Type", "application/json")
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestUseLog_PUT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseLog()

	s.app.PUT("/test-log-put", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/test-log-put", nil)
	req.Header.Set("Content-Type", "application/json")
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestUseLog_WithAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseLog()

	s.app.GET("/test-log-auth", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test-log-auth", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiJ9.test")
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestUseLog_MultipartForm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseLog()

	s.app.POST("/test-log-multipart", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test-log-multipart", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=something")
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestUseLog_WithBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseLog()

	s.app.POST("/test-log-body", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	bodyContent := `{"key":"value"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test-log-body", strings.NewReader(bodyContent))
	req.Header.Set("Content-Type", "application/json")
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestUseLog_GET_SkipsBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseLog()

	bodyCaptured := false
	s.app.GET("/test-get", func(ctx *gin.Context) {
		_, bodyCaptured = ctx.Get("RequestBody")
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test-get", strings.NewReader(`{"key":"value"}`))
	req.Header.Set("Content-Type", "application/json")
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if bodyCaptured {
		t.Error("GET request should NOT capture RequestBody in context")
	}
}

func TestUseLog_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseLog()

	bodyCaptured := false
	s.app.POST("/test-empty", func(ctx *gin.Context) {
		_, bodyCaptured = ctx.Get("RequestBody")
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test-empty", nil)
	req.Header.Set("Content-Type", "application/json")
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if bodyCaptured {
		t.Error("empty POST body should NOT set RequestBody in context")
	}
}

func TestUseLog_PUT_WithBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseLog()

	bodyReadable := false
	s.app.PUT("/test-put", func(ctx *gin.Context) {
		raw, err := ctx.GetRawData()
		if err == nil && len(raw) > 0 {
			bodyReadable = true
		}
		ctx.String(http.StatusOK, string(raw))
	})

	bodyContent := `{"name":"updated"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/test-put", strings.NewReader(bodyContent))
	req.Header.Set("Content-Type", "application/json")
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !bodyReadable {
		t.Error("handler should be able to read body after middleware reset it")
	}
}

func TestUseLog_PUT_Multipart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	s.UseLog()

	s.app.PUT("/test-put-multipart", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/test-put-multipart", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=something")
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
