package miniprogram

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"resty.dev/v3"
)

func newTestMiniProgram(t *testing.T, srv *httptest.Server) *MiniProgram {
	t.Helper()
	return &MiniProgram{
		config:     &Config{AppID: "appid", Secret: "secret"},
		httpClient: resty.New().SetBaseURL(srv.URL),
		token:      &Token{mutex: &sync.Mutex{}},
	}
}

func TestGetAccessTokenCacheHit(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"access_token":"tok1","expires_in":7200}`))
	}))
	defer srv.Close()

	wx := newTestMiniProgram(t, srv)
	ctx := context.Background()

	first, err := wx.GetAccessToken(ctx, false)
	if err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", calls.Load())
	}
	// 命中/刷新两条路径都返回预留 200 秒后的值
	if first.ExpiresIn != 7000 {
		t.Fatalf("expected adjusted expires_in 7000, got %d", first.ExpiresIn)
	}

	second, err := wx.GetAccessToken(ctx, false)
	if err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("cache should be hit, got %d calls", calls.Load())
	}
	if second.AccessToken != "tok1" || second.ExpiresIn != 7000 {
		t.Fatalf("unexpected cached resp: %+v", second)
	}
}

func TestGetAccessTokenForceRefresh(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"access_token":"tok1","expires_in":7200}`))
	}))
	defer srv.Close()

	wx := newTestMiniProgram(t, srv)
	ctx := context.Background()

	if _, err := wx.GetAccessToken(ctx, false); err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	if _, err := wx.GetAccessToken(ctx, true); err != nil {
		t.Fatalf("GetAccessToken(forceRefresh): %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 calls with force refresh, got %d", calls.Load())
	}
}

func TestGetAccessTokenExpired(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"access_token":"tok1","expires_in":7200}`))
	}))
	defer srv.Close()

	wx := newTestMiniProgram(t, srv)
	ctx := context.Background()

	if _, err := wx.GetAccessToken(ctx, false); err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	// 模拟 token 过期(缓存有效期 7000 秒)
	wx.token.NetTime = time.Now().Unix() - 8000
	if _, err := wx.GetAccessToken(ctx, false); err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 calls after expiry, got %d", calls.Load())
	}
}

func TestGetAccessTokenAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40164,"errmsg":"invalid ip"}`))
	}))
	defer srv.Close()

	wx := newTestMiniProgram(t, srv)
	_, err := wx.GetAccessToken(context.Background(), false)
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != 40164 {
		t.Fatalf("unexpected errcode: %d", apiErr.ErrCode)
	}
}
