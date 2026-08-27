package douyin

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/alibabacloud-go/tea/tea"
)

// newTextCensorDouyin 构造预置 client_token 缓存的 Douyin,避免测试走官方 SDK 网络调用
func newTextCensorDouyin(srvURL string) *Douyin {
	return &Douyin{
		config: &Config{AppID: "appid", TextCensorURL: srvURL},
		token:  &Token{mutex: &sync.Mutex{}, NetTime: time.Now().Unix(), ExpiresIn: 7000, AccessToken: "tok"},
	}
}

func TestContentSecurityTextDetectSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("access-token") != "tok" {
			t.Errorf("unexpected access-token header: %q", r.Header.Get("access-token"))
		}
		var req TextCensorRequest
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.AppID != "appid" || len(req.Content) != 1 || req.Content[0] != "hello world" {
			t.Errorf("unexpected body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"log_id":"log1","err_no":0,"data":[{"code":0,"msg":"ok","data_id":"d1","task_id":"t1","predicts":[{"target":"risk","model_name":"risk_model","prob":0.99,"hit":true}]}]}`))
	}))
	defer srv.Close()

	resp, err := newTextCensorDouyin(srv.URL).ContentSecurityTextDetect(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.LogID != "log1" || len(resp.Data) != 1 || !resp.Data[0].Predicts[0].Hit {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestContentSecurityTextDetectRequestLevelError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"err_no":20000,"err_msg":"invalid token"}`))
	}))
	defer srv.Close()

	_, err := newTextCensorDouyin(srv.URL).ContentSecurityTextDetect(context.Background(), "x")
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if !errors.Is(err, ErrBusiness) {
		t.Fatalf("expected ErrBusiness, got %v", err)
	}
	if apiErr.ErrCode != 20000 || apiErr.ErrMsg != "invalid token" {
		t.Fatalf("unexpected err: %+v", apiErr)
	}
}

func TestContentSecurityTextDetectTaskLevelError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"err_no":0,"data":[{"code":1001,"msg":"task failed","data_id":"d1","task_id":"t1"}]}`))
	}))
	defer srv.Close()

	_, err := newTextCensorDouyin(srv.URL).ContentSecurityTextDetect(context.Background(), "x")
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != 1001 || apiErr.ErrMsg != "task failed" {
		t.Fatalf("unexpected err: %+v", apiErr)
	}
}

func TestContentSecurityTextDetectEmptyContents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be sent")
	}))
	defer srv.Close()

	_, err := newTextCensorDouyin(srv.URL).ContentSecurityTextDetect(context.Background())
	if err == nil {
		t.Fatal("expected error for empty contents")
	}
	if !errors.Is(err, ErrLocal) {
		t.Fatalf("expected ErrLocal, got %v", err)
	}
}

func TestContentSecurityTextDetectHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer srv.Close()

	_, err := newTextCensorDouyin(srv.URL).ContentSecurityTextDetect(context.Background(), "x")
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if !errors.Is(err, ErrHTTP) {
		t.Fatalf("expected ErrHTTP, got %v", err)
	}
	if apiErr.Status != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d", apiErr.Status)
	}
}

func TestContentSecurityTextSafe(t *testing.T) {
	srvHit := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"err_no":0,"data":[{"code":0,"msg":"ok","predicts":[{"model_name":"risk","hit":true}]}]}`))
	}))
	defer srvHit.Close()
	if safe, err := newTextCensorDouyin(srvHit.URL).ContentSecurityTextSafe(context.Background(), "bad"); err != nil || safe {
		t.Fatalf("expected unsafe: safe=%v err=%v", safe, err)
	}

	srvClean := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"err_no":0,"data":[{"code":0,"msg":"ok","predicts":[{"model_name":"risk","hit":false}]}]}`))
	}))
	defer srvClean.Close()
	if safe, err := newTextCensorDouyin(srvClean.URL).ContentSecurityTextSafe(context.Background(), "good"); err != nil || !safe {
		t.Fatalf("expected safe: safe=%v err=%v", safe, err)
	}
}

func TestParseSDKErrorCode(t *testing.T) {
	code, kind := parseSDKErrorCode(&tea.SDKError{Code: tea.String("40001"), Message: tea.String("x")})
	if code != 40001 || kind != ErrKindBusiness {
		t.Fatalf("expected 40001/business, got %d/%v", code, kind)
	}
	code, kind = parseSDKErrorCode(&tea.SDKError{Code: tea.String("abc"), Message: tea.String("x")})
	if code != ECodeCall || kind != ErrKindLocal {
		t.Fatalf("expected ECodeCall/local for non-numeric code, got %d/%v", code, kind)
	}
	code, kind = parseSDKErrorCode(&tea.SDKError{})
	if code != ECodeCall || kind != ErrKindLocal {
		t.Fatalf("expected ECodeCall/local for empty code, got %d/%v", code, kind)
	}
}
