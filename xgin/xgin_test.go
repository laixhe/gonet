package xgin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zaptest"
)

func TestMaskAuth(t *testing.T) {
	tests := []struct {
		name string
		auth string
		want string
	}{
		{"empty", "", ""},
		{"no space", "abc123", "***"},
		{"bearer token", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "Bearer ***"},
		{"basic auth", "Basic dXNlcjpwYXNz", "Basic ***"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maskAuth(tt.auth); got != tt.want {
				t.Errorf("maskAuth(%q) = %q, want %q", tt.auth, got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Run("debug mode", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		s := New(true, logger)
		if s == nil {
			t.Fatal("New() returned nil")
		}
		if !s.isDebug {
			t.Error("isDebug should be true")
		}
		if s.logger != logger {
			t.Error("logger mismatch")
		}
		if s.app == nil {
			t.Error("app should not be nil")
		}
	})

	t.Run("release mode", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		s := New(false, logger)
		if s == nil {
			t.Fatal("New() returned nil")
		}
		if s.isDebug {
			t.Error("isDebug should be false")
		}
	})
}

func TestServer_App(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger)
	app := s.App()
	if app == nil {
		t.Fatal("App() returned nil")
	}
	if app != s.app {
		t.Error("App() should return the same *gin.Engine")
	}
}

func TestServer_ListenError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(false, logger)

	err := s.Listen(":99999")
	if err == nil {
		t.Error("expected error for invalid port")
		s.app = nil
	}
}

func TestServer_NewWithOpts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger, func(engine *gin.Engine) {
		engine.MaxMultipartMemory = 8 << 20
	})
	if s == nil {
		t.Fatal("New() with opts returned nil")
	}
	if s.app.MaxMultipartMemory != 8<<20 {
		t.Errorf("MaxMultipartMemory = %d, want %d", s.app.MaxMultipartMemory, 8<<20)
	}
}

func TestServer_New_WithNilLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := New(false, nil)
	if s == nil {
		t.Fatal("New() with nil logger returned nil")
	}
	if s.logger != nil {
		t.Error("logger should be nil")
	}
}

// TestFullChain 测试多条中间件同时注册时互不干扰，请求正常完成。
func TestFullChain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)
	s := New(true, logger).
		UseLog().
		UseLimiter(LimiterConfig{Rate: 100, Burst: 200}).
		UseTimeout(1 * time.Second).
		UseCors().
		UseRecover()

	s.app.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	s.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
