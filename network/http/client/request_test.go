package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/laixhe/gonet/network/header"
)

// echo 回显的请求信息
type echo struct {
	Method      string
	Path        string
	Query       string
	Header      http.Header
	Body        string
	ContentType string
}

// newEchoServer 创建回显请求信息的测试服务器
func newEchoServer(t *testing.T, last *echo) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*last = echo{
			Method:      r.Method,
			Path:        r.URL.Path,
			Query:       r.URL.RawQuery,
			Header:      r.Header.Clone(),
			Body:        string(b),
			ContentType: r.Header.Get(header.ContentType),
		}
		_, _ = w.Write(b) // 回显 body
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestRequestText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello:" + r.URL.Query().Get("name")))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	body, err := c.Get("/greet").SetQueryParam("name", "world").Text()
	if err != nil {
		t.Fatalf("Text 失败: %v", err)
	}
	if body != "hello:world" {
		t.Errorf("body = %q, want %q", body, "hello:world")
	}
}

func TestRequestMethods(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)
	c := NewClient(base)

	methods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions,
	}
	for _, m := range methods {
		_, err := c.newRequest(m, "/path").Text()
		if err != nil {
			t.Fatalf("%s 请求失败: %v", m, err)
		}
		if last.Method != m {
			t.Errorf("method = %q, want %q", last.Method, m)
		}
		if last.Path != "/path" {
			t.Errorf("path = %q, want /path", last.Path)
		}
	}
}

func TestRequestQueryParams(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	_, err := NewClient(base).Get("/path").
		SetQueryParam("a", "1").
		SetQueryParams(map[string]string{"b": "2", "c": "3"}).
		Text()
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	// url.Values.Encode 按键名排序, 结果确定
	if last.Query != "a=1&b=2&c=3" {
		t.Errorf("query = %q, want %q", last.Query, "a=1&b=2&c=3")
	}
}

func TestRequestHeaders(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	_, err := NewClient(base).Get("/header").
		SetHeader(header.XRequestID, "req-1").
		SetHeaders(map[string]string{"X-Trace": "trace-1"}).
		SetUserAgent("test-agent").
		Text()
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	if got := last.Header.Get(header.XRequestID); got != "req-1" {
		t.Errorf("X-Request-Id = %q, want req-1", got)
	}
	if got := last.Header.Get("X-Trace"); got != "trace-1" {
		t.Errorf("X-Trace = %q, want trace-1", got)
	}
	if got := last.Header.Get(header.UserAgent); got != "test-agent" {
		t.Errorf("User-Agent = %q, want test-agent", got)
	}
}

func TestRequestJSON(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	var out map[string]string
	err := NewClient(base).Post("/json").SetJSON(map[string]string{"name": "gonet"}).JSON(&out)
	if err != nil {
		t.Fatalf("JSON 请求失败: %v", err)
	}
	if out["name"] != "gonet" {
		t.Errorf("out = %v, want name=gonet", out)
	}
	if last.Method != http.MethodPost {
		t.Errorf("method = %q, want POST", last.Method)
	}
	if last.ContentType != header.ApplicationJSONCharsetUTF8 {
		t.Errorf("Content-Type = %q, want %q", last.ContentType, header.ApplicationJSONCharsetUTF8)
	}
}

func TestRequestJSONWithResult(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	var out map[string]string
	err := NewClient(base).Post("/json").
		SetJSON(map[string]string{"k": "v"}).
		SetResult(&out).
		JSON(nil)
	if err != nil {
		t.Fatalf("JSON 请求失败: %v", err)
	}
	if out["k"] != "v" {
		t.Errorf("out = %v, want k=v", out)
	}
}

func TestRequestJSONNilTarget(t *testing.T) {
	base := newEchoServer(t, &echo{})
	err := NewClient(base).Get("/json").JSON(nil)
	if err == nil {
		t.Fatal("期望解码目标为空的错误")
	}
}

func TestRequestJSONStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	var out any
	err := NewClient(srv.URL).Get("/404").JSON(&out)
	var he *HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("期望 *HTTPError, 实际 %T: %v", err, err)
	}
	if he.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", he.StatusCode, http.StatusNotFound)
	}
	if !strings.Contains(he.Error(), "404") {
		t.Errorf("Error() = %q, 应包含状态码", he.Error())
	}
}

func TestRequestSetBody(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	// []byte
	if _, err := NewClient(base).Post("/raw").SetBody([]byte("raw-data")).Text(); err != nil {
		t.Fatalf("[]byte body 失败: %v", err)
	}
	if last.Body != "raw-data" {
		t.Errorf("body = %q, want raw-data", last.Body)
	}

	// string
	if _, err := NewClient(base).Post("/str").SetBody("str-data").Text(); err != nil {
		t.Fatalf("string body 失败: %v", err)
	}
	if last.Body != "str-data" {
		t.Errorf("body = %q, want str-data", last.Body)
	}

	// 其他类型按 JSON 序列化
	if _, err := NewClient(base).Post("/obj").SetBody(map[string]string{"k": "v"}).Text(); err != nil {
		t.Fatalf("object body 失败: %v", err)
	}
	if last.ContentType != header.ApplicationJSONCharsetUTF8 {
		t.Errorf("Content-Type = %q, want %q", last.ContentType, header.ApplicationJSONCharsetUTF8)
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(last.Body), &m); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if m["k"] != "v" {
		t.Errorf("body json = %v, want k=v", m)
	}
}

func TestRequestForm(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	_, err := NewClient(base).Post("/form").SetFormData(map[string]string{"a": "1", "b": "2"}).Text()
	if err != nil {
		t.Fatalf("表单请求失败: %v", err)
	}
	if last.ContentType != header.ApplicationForm {
		t.Errorf("Content-Type = %q, want %q", last.ContentType, header.ApplicationForm)
	}
	if last.Body != "a=1&b=2" {
		t.Errorf("body = %q, want a=1&b=2", last.Body)
	}
}

func TestRequestMultipart(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	_, err := NewClient(base).Post("/upload").
		SetFormData(map[string]string{"field": "value"}).
		SetFile("file", "a.txt", []byte("hello")).
		Text()
	if err != nil {
		t.Fatalf("multipart 请求失败: %v", err)
	}
	if !strings.HasPrefix(last.ContentType, "multipart/form-data; boundary=") {
		t.Errorf("Content-Type = %q, 应为 multipart/form-data", last.ContentType)
	}
	if !strings.Contains(last.Body, `name="field"`) || !strings.Contains(last.Body, "value") {
		t.Errorf("body 缺少表单字段: %q", last.Body)
	}
	if !strings.Contains(last.Body, `filename="a.txt"`) || !strings.Contains(last.Body, "hello") {
		t.Errorf("body 缺少文件内容: %q", last.Body)
	}
}

func TestSetFileFromPath(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.txt")
	if err := os.WriteFile(p, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	var last echo
	base := newEchoServer(t, &last)
	_, err := NewClient(base).Post("/up").SetFileFromPath("file", p).Text()
	if err != nil {
		t.Fatalf("multipart 请求失败: %v", err)
	}
	if !strings.Contains(last.Body, `filename="x.txt"`) || !strings.Contains(last.Body, "content") {
		t.Errorf("body 缺少文件: %q", last.Body)
	}
}

func TestRequestAuth(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	// Basic
	if _, err := NewClient(base).Get("/auth").SetBasicAuth("user", "pass").Text(); err != nil {
		t.Fatalf("Basic 请求失败: %v", err)
	}
	if got := last.Header.Get(header.Authorization); got != "Basic dXNlcjpwYXNz" {
		t.Errorf("Authorization = %q, want Basic dXNlcjpwYXNz", got)
	}

	// Bearer
	if _, err := NewClient(base).Get("/auth").SetBearerToken("token-1").Text(); err != nil {
		t.Fatalf("Bearer 请求失败: %v", err)
	}
	if got := last.Header.Get(header.Authorization); got != "Bearer token-1" {
		t.Errorf("Authorization = %q, want Bearer token-1", got)
	}
}

func TestRequestCookie(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	_, err := NewClient(base).Get("/cookie").
		SetCookie(&http.Cookie{Name: "sid", Value: "abc"}).
		Text()
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	if got := last.Header.Get(header.Cookie); got != "sid=abc" {
		t.Errorf("Cookie = %q, want sid=abc", got)
	}
}

func TestClientHooks(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	var beforeCalled, afterCalled bool
	c := NewClient(base)
	c.OnBeforeRequest(func(req *http.Request) error {
		beforeCalled = true
		req.Header.Set("X-Hook", "1")
		return nil
	})
	c.OnAfterResponse(func(resp *http.Response) error {
		afterCalled = true
		return nil
	})

	if _, err := c.Get("/hook").Text(); err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	if !beforeCalled || !afterCalled {
		t.Errorf("钩子未执行: before=%v after=%v", beforeCalled, afterCalled)
	}
	if got := last.Header.Get("X-Hook"); got != "1" {
		t.Errorf("X-Hook = %q, want 1", got)
	}
}

func TestBeforeHookError(t *testing.T) {
	c := NewClient("")
	wantErr := errors.New("stop")
	c.OnBeforeRequest(func(*http.Request) error { return wantErr })

	_, err := c.Get("/x").Text()
	if !errors.Is(err, wantErr) {
		t.Fatalf("期望钩子错误 %v, 实际 %v", wantErr, err)
	}
}

func TestAfterHookError(t *testing.T) {
	base := newEchoServer(t, &echo{})
	wantErr := errors.New("after-stop")
	c := NewClient(base)
	c.OnAfterResponse(func(*http.Response) error { return wantErr })

	_, err := c.Get("/x").Text()
	if !errors.Is(err, wantErr) {
		t.Fatalf("期望钩子错误 %v, 实际 %v", wantErr, err)
	}
}

func TestClientTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer srv.Close()

	c := NewClient(srv.URL).SetTimeout(50 * time.Millisecond)
	if _, err := c.Get("/slow").Text(); err == nil {
		t.Fatal("期望超时错误")
	}
}

func TestRequestContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewClient(srv.URL).Get("/x").SetContext(ctx).Text(); err == nil {
		t.Fatal("期望上下文取消错误")
	}
}

func TestSetBaseURL(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	c := NewClient("").SetBaseURL(base)
	if _, err := c.Get("/greet").Text(); err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	if last.Path != "/greet" {
		t.Errorf("path = %q, want /greet", last.Path)
	}
}

func TestJoinURL(t *testing.T) {
	cases := []struct{ base, ref, want string }{
		{"https://x.com", "/a", "https://x.com/a"},
		{"https://x.com/", "/a", "https://x.com/a"},
		{"https://x.com/api", "b", "https://x.com/api/b"},
		{"https://x.com/api/", "/b", "https://x.com/api/b"},
		{"", "/a", "/a"},
		{"https://x.com", "https://other.com/a", "https://other.com/a"},
	}
	for _, tc := range cases {
		if got := joinURL(tc.base, tc.ref); got != tc.want {
			t.Errorf("joinURL(%q, %q) = %q, want %q", tc.base, tc.ref, got, tc.want)
		}
	}
}

// ======================== 自动重试 ========================

func TestRequestRetry(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	body, err := NewClient(srv.URL).
		Get("/unstable").
		SetRetryCount(3).
		SetRetryWaitTime(time.Millisecond).
		Text()
	if err != nil {
		t.Fatalf("重试后仍失败: %v", err)
	}
	if body != "ok" {
		t.Errorf("body = %q, want ok", body)
	}
}

func TestRequestRetryExhausted(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL).
		Get("/fail").
		SetRetryCount(2).
		SetRetryWaitTime(time.Millisecond).
		SetCheckStatus(true).
		Text()
	if err == nil {
		t.Fatal("期望重试耗尽后错误")
	}
	var he *HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("期望 *HTTPError, 实际 %T: %v", err, err)
	}
	if he.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want 500", he.StatusCode)
	}
	if got := n.Load(); got != 3 {
		t.Errorf("请求次数 = %d, want 3", got)
	}
}

func TestRequestRetryCondition(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) == 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	body, err := NewClient(srv.URL).
		Get("/cond").
		SetRetryCount(2).
		SetRetryWaitTime(time.Millisecond).
		AddRetryCondition(RetryConditionStatus(http.StatusBadRequest)).
		Text()
	if err != nil {
		t.Fatalf("自定义条件重试失败: %v", err)
	}
	if body != "ok" {
		t.Errorf("body = %q, want ok", body)
	}
}

func TestClientRetryInherit(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL).SetRetryCount(3).SetRetryWaitTime(time.Millisecond)
	body, err := c.Get("/x").Text()
	if err != nil {
		t.Fatalf("继承客户端重试后仍失败: %v", err)
	}
	if body != "ok" {
		t.Errorf("body = %q, want ok", body)
	}
}

func TestRequestRetryDisabled(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient(srv.URL).SetRetryCount(3)
	_, err := c.Get("/x").SetRetryCount(-1).SetCheckStatus(true).Text()
	if err == nil {
		t.Fatal("期望状态码错误")
	}
	if got := n.Load(); got != 1 {
		t.Errorf("请求次数 = %d, want 1", got)
	}
}

func TestRetryWithReaderBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL).
		Post("/x").
		SetBody(strings.NewReader("data")).
		SetRetryCount(2).
		Text()
	if err == nil || !strings.Contains(err.Error(), "不支持重试") {
		t.Fatalf("期望 io.Reader 不可重试错误, 实际 %v", err)
	}
}

// ======================== 状态码校验 ========================

func TestRequestCheckStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadRequest)
	}))
	defer srv.Close()

	// 默认不校验, 裸返回
	body, err := NewClient(srv.URL).Get("/x").Text()
	if err != nil {
		t.Fatalf("默认不应报错: %v", err)
	}
	if !strings.Contains(body, "boom") {
		t.Errorf("body = %q, want 包含 boom", body)
	}

	// 启用校验后返回 HTTPError
	_, err = NewClient(srv.URL).Get("/x").SetCheckStatus(true).Text()
	var he *HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("期望 *HTTPError, 实际 %T: %v", err, err)
	}
	if he.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", he.StatusCode)
	}
}

func TestClientCheckStatusInherit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadRequest)
	}))
	defer srv.Close()

	c := NewClient(srv.URL).SetCheckStatus(true)
	_, err := c.Get("/x").Text()
	var he *HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("期望 *HTTPError, 实际 %T: %v", err, err)
	}
	// 请求级可关闭
	body, err := c.Get("/x").SetCheckStatus(false).Text()
	if err != nil {
		t.Fatalf("请求级关闭校验后不应报错: %v", err)
	}
	if !strings.Contains(body, "boom") {
		t.Errorf("body = %q, want 包含 boom", body)
	}
}

// ======================== 下载到文件 ========================

func TestRequestDownload(t *testing.T) {
	want := strings.Repeat("data", 1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(want))
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "sub", "file.bin")
	if err := NewClient(srv.URL).Get("/download").Download(path); err != nil {
		t.Fatalf("Download 失败: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Errorf("文件内容长度 = %d, want %d", len(data), len(want))
	}
}

func TestRequestDownloadStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "denied", http.StatusForbidden)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "x.bin")
	err := NewClient(srv.URL).Get("/x").SetCheckStatus(true).Download(path)
	var he *HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("期望 *HTTPError, 实际 %T: %v", err, err)
	}
	if he.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want 403", he.StatusCode)
	}
}

// ======================== 错误体解码 ========================

// apiError 测试用错误结构体
type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *apiError) Error() string { return fmt.Sprintf("code=%d %s", e.Code, e.Message) }

func TestRequestSetError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":1001,"message":"bad request"}`))
	}))
	defer srv.Close()

	var out any
	err := NewClient(srv.URL).Get("/error").SetError(&apiError{}).JSON(&out)
	var ae *apiError
	if !errors.As(err, &ae) {
		t.Fatalf("期望 *apiError, 实际 %T: %v", err, err)
	}
	if ae.Code != 1001 || ae.Message != "bad request" {
		t.Errorf("apiError = %+v, want code=1001 message=bad request", ae)
	}
}

// ======================== 响应体大小上限 ========================

func TestRequestMaxBodySize(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("0123456789"))
	}))
	defer srv.Close()

	if _, err := NewClient(srv.URL).Get("/x").SetMaxBodySize(5).Bytes(); err == nil {
		t.Fatal("期望超出大小限制错误")
	}
	// 未超限正常返回
	data, err := NewClient(srv.URL).Get("/x").SetMaxBodySize(100).Bytes()
	if err != nil {
		t.Fatalf("未超限不应报错: %v", err)
	}
	if string(data) != "0123456789" {
		t.Errorf("body = %q", data)
	}
}

// ======================== 流式 multipart ========================

func TestRequestSetFileStream(t *testing.T) {
	var last echo
	base := newEchoServer(t, &last)

	_, err := NewClient(base).Post("/upload").
		SetFormData(map[string]string{"field": "value"}).
		SetFileStream("file", "big.bin", strings.NewReader("stream-content")).
		Text()
	if err != nil {
		t.Fatalf("流式上传失败: %v", err)
	}
	if !strings.Contains(last.Body, `filename="big.bin"`) {
		t.Errorf("body 缺少文件名: %q", last.Body)
	}
	if !strings.Contains(last.Body, "stream-content") {
		t.Errorf("body 缺少流内容: %q", last.Body)
	}
	if !strings.Contains(last.Body, `name="field"`) || !strings.Contains(last.Body, "value") {
		t.Errorf("body 缺少表单字段: %q", last.Body)
	}
}

// ======================== 客户端级默认配置 ========================

func TestClientConfigDefaults(t *testing.T) {
	c := NewClient("").
		SetCheckStatus(true).
		SetMaxBodySize(100).
		SetRetryCount(2).
		SetRetryWaitTime(5 * time.Millisecond).
		SetRetryMaxWaitTime(10 * time.Millisecond)

	r := c.Get("/x")
	if !r.shouldCheckStatus() {
		t.Error("shouldCheckStatus = false, want true")
	}
	if r.bodyLimit() != 100 {
		t.Errorf("bodyLimit = %d, want 100", r.bodyLimit())
	}
	if r.retries() != 2 {
		t.Errorf("retries = %d, want 2", r.retries())
	}
	if r.retryWaitTime() != 5*time.Millisecond {
		t.Errorf("retryWaitTime = %v, want 5ms", r.retryWaitTime())
	}
	if r.retryMaxWaitTime() != 10*time.Millisecond {
		t.Errorf("retryMaxWaitTime = %v, want 10ms", r.retryMaxWaitTime())
	}

	// 请求级覆盖
	r.SetCheckStatus(false)
	if r.shouldCheckStatus() {
		t.Error("请求级关闭校验后 shouldCheckStatus 仍为 true")
	}
}
