package apiutil

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"resty.dev/v3"
)

func newTestClient(t *testing.T, srv *httptest.Server) *resty.Client {
	t.Helper()
	return resty.New().SetBaseURL(srv.URL)
}

type fakeResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	Data    string `json:"data"`
}

func TestGetSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ok" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("a"); got != "1" {
			t.Errorf("unexpected query a=%s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"data":"hello"}`))
	}))
	defer srv.Close()

	resp, err := Get[fakeResp](context.Background(), newTestClient(t, srv), "/ok", map[string]string{"a": "1"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if resp.Data != "hello" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestGetBusinessError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
	}))
	defer srv.Close()

	_, err := Get[fakeResp](context.Background(), newTestClient(t, srv), "/err", nil)
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != 40029 || apiErr.ErrMsg != "invalid code" {
		t.Fatalf("unexpected err: %+v", apiErr)
	}
}

func TestGetHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer srv.Close()

	_, err := Get[fakeResp](context.Background(), newTestClient(t, srv), "/err", nil)
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.Status != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d", apiErr.Status)
	}
}

func TestPostBinaryImage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte{1, 2, 3})
	}))
	defer srv.Close()

	resp, err := PostBinary(context.Background(), newTestClient(t, srv), "/img", nil, nil)
	if err != nil {
		t.Fatalf("PostBinary: %v", err)
	}
	if resp.ContentType != "image/jpeg" || len(resp.Data) != 3 || resp.Data[0] != 1 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestPostBinaryJSONError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40129,"errmsg":"scene invalid"}`))
	}))
	defer srv.Close()

	_, err := PostBinary(context.Background(), newTestClient(t, srv), "/img", nil, nil)
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != 40129 {
		t.Fatalf("unexpected errcode: %d", apiErr.ErrCode)
	}
}

func TestContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := Get[fakeResp](ctx, newTestClient(t, srv), "/slow", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ApiError, got %v", err)
	}
	if apiErr.ErrCode != -1 {
		t.Fatalf("unexpected errcode: %d", apiErr.ErrCode)
	}
}
