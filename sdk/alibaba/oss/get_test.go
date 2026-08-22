package oss

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// roundTripFunc 自定义 RoundTripper,拦截并返回固定响应
type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// newGetInfoOClient 构造测试用 OClient
func newGetInfoOClient() *OClient {
	return &OClient{config: &Config{Bucket: "test", Region: "cn-shenzhen"}}
}

// TestGetInfo 验证 GetInfo 请求 URL 与响应解析
func TestGetInfo(t *testing.T) {
	oc := newGetInfoOClient()
	old := imageInfoHTTPClient
	defer func() { imageInfoHTTPClient = old }()
	imageInfoHTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if !strings.Contains(r.URL.String(), "?x-oss-process=image/info") {
			t.Errorf("unexpected url: %s", r.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"FileSize":{"value":"100"},"Format":{"value":"jpg"},"ImageHeight":{"value":"10"},"ImageWidth":{"value":"20"}}`)),
		}, nil
	})}

	info, err := oc.GetInfo(context.Background(), "a.jpg")
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info.FileSize.Value != "100" || info.Format.Value != "jpg" || info.ImageHeight.Value != "10" || info.ImageWidth.Value != "20" {
		t.Fatalf("unexpected info: %+v", info)
	}
}

// TestGetInfoHTTPError 验证非 200 状态码返回错误
func TestGetInfoHTTPError(t *testing.T) {
	oc := newGetInfoOClient()
	old := imageInfoHTTPClient
	defer func() { imageInfoHTTPClient = old }()
	imageInfoHTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusForbidden, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
	})}

	if _, err := oc.GetInfo(context.Background(), "a.jpg"); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

// TestGetInfoTimeout 验证对端不响应时超时返回错误(而非永久挂起)
func TestGetInfoTimeout(t *testing.T) {
	oc := newGetInfoOClient()
	old := imageInfoHTTPClient
	defer func() { imageInfoHTTPClient = old }()
	imageInfoHTTPClient = &http.Client{
		Timeout: 50 * time.Millisecond,
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			time.Sleep(300 * time.Millisecond)
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
		}),
	}

	if _, err := oc.GetInfo(context.Background(), "a.jpg"); err == nil {
		t.Fatal("expected timeout error")
	}
}
