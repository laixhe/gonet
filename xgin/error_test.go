package xgin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestError_Error(t *testing.T) {
	e := &Error{Code: 500, Message: "test error"}
	if e.Error() != "test error" {
		t.Errorf("Error() = %q, want %q", e.Error(), "test error")
	}
}

func TestServerError(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := ServerError()
		if e.Code != http.StatusInternalServerError {
			t.Errorf("Code = %d, want %d", e.Code, http.StatusInternalServerError)
		}
		if e.Message != "Internal Server Error" {
			t.Errorf("Message = %q, want %q", e.Message, "Internal Server Error")
		}
	})

	t.Run("custom message", func(t *testing.T) {
		e := ServerError("custom server error")
		if e.Code != http.StatusInternalServerError {
			t.Errorf("Code = %d, want %d", e.Code, http.StatusInternalServerError)
		}
		if e.Message != "custom server error" {
			t.Errorf("Message = %q, want %q", e.Message, "custom server error")
		}
	})

	t.Run("multiple messages uses first", func(t *testing.T) {
		e := ServerError("first", "second")
		if e.Message != "first" {
			t.Errorf("Message = %q, want %q", e.Message, "first")
		}
	})
}

func TestAuthorizedError(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := AuthorizedError()
		if e.Code != http.StatusUnauthorized {
			t.Errorf("Code = %d, want %d", e.Code, http.StatusUnauthorized)
		}
		if e.Message != "Unauthorized" {
			t.Errorf("Message = %q, want %q", e.Message, "Unauthorized")
		}
	})

	t.Run("custom message", func(t *testing.T) {
		e := AuthorizedError("custom unauthorized")
		if e.Code != http.StatusUnauthorized {
			t.Errorf("Code = %d, want %d", e.Code, http.StatusUnauthorized)
		}
		if e.Message != "custom unauthorized" {
			t.Errorf("Message = %q, want %q", e.Message, "custom unauthorized")
		}
	})
}

func TestParamError(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := ParamError()
		if e.Code != http.StatusUnprocessableEntity {
			t.Errorf("Code = %d, want %d", e.Code, http.StatusUnprocessableEntity)
		}
		if e.Message != "Param Error" {
			t.Errorf("Message = %q, want %q", e.Message, "Param Error")
		}
	})

	t.Run("custom message", func(t *testing.T) {
		e := ParamError("custom param error")
		if e.Code != http.StatusUnprocessableEntity {
			t.Errorf("Code = %d, want %d", e.Code, http.StatusUnprocessableEntity)
		}
		if e.Message != "custom param error" {
			t.Errorf("Message = %q, want %q", e.Message, "custom param error")
		}
	})
}

func TestTipError(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := TipError()
		if e.Code != 400 {
			t.Errorf("Code = %d, want 400", e.Code)
		}
		if e.Message != "Tip Error" {
			t.Errorf("Message = %q, want %q", e.Message, "Tip Error")
		}
	})

	t.Run("custom message", func(t *testing.T) {
		e := TipError("余额不足")
		if e.Code != 400 {
			t.Errorf("Code = %d, want 400", e.Code)
		}
		if e.Message != "余额不足" {
			t.Errorf("Message = %q, want %q", e.Message, "余额不足")
		}
	})
}

func TestErrorRecoveryFunc(t *testing.T) {
	if ErrorRecoveryFunc == nil {
		t.Fatal("ErrorRecoveryFunc should not be nil")
	}
}

func TestErrorRecoveryFunc_Behavior(t *testing.T) {
	// 验证 ErrorRecoveryFunc 实际写入 500 JSON 响应（ServerError 格式）
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	ErrorRecoveryFunc(ctx, "test panic")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var resp Error
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Code != http.StatusInternalServerError {
		t.Errorf("resp.Code = %d, want %d", resp.Code, http.StatusInternalServerError)
	}
	if resp.Message != "Internal Server Error" {
		t.Errorf("resp.Message = %q, want %q", resp.Message, "Internal Server Error")
	}
}

func TestTimeoutError(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := TimeoutError()
		if e.Code != http.StatusRequestTimeout {
			t.Errorf("Code = %d, want %d", e.Code, http.StatusRequestTimeout)
		}
		if e.Message != "Request Timeout" {
			t.Errorf("Message = %q, want %q", e.Message, "Request Timeout")
		}
	})
	t.Run("custom message", func(t *testing.T) {
		e := TimeoutError("超时")
		if e.Message != "超时" {
			t.Errorf("Message = %q, want %q", e.Message, "超时")
		}
	})
}

func TestNotFoundError(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := NotFoundError()
		if e.Code != http.StatusNotFound {
			t.Errorf("Code = %d, want %d", e.Code, http.StatusNotFound)
		}
		if e.Message != "Not Found" {
			t.Errorf("Message = %q, want %q", e.Message, "Not Found")
		}
	})
	t.Run("custom message", func(t *testing.T) {
		e := NotFoundError("资源不存在")
		if e.Message != "资源不存在" {
			t.Errorf("Message = %q, want %q", e.Message, "资源不存在")
		}
	})
}

func TestRateLimitError(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := RateLimitError()
		if e.Code != http.StatusTooManyRequests {
			t.Errorf("Code = %d, want %d", e.Code, http.StatusTooManyRequests)
		}
		if e.Message != "Too Many Requests" {
			t.Errorf("Message = %q, want %q", e.Message, "Too Many Requests")
		}
	})
	t.Run("custom message", func(t *testing.T) {
		e := RateLimitError("请求过于频繁")
		if e.Message != "请求过于频繁" {
			t.Errorf("Message = %q, want %q", e.Message, "请求过于频繁")
		}
	})
}
