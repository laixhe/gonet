// Package client 提供链式 API 的 HTTP 客户端, 内置连接池配置与请求辅助方法。
//
// 用法:
//
//	c := client.NewClient("https://api.example.com").SetTimeout(5 * time.Second)
//	body, err := c.Post("/login").
//		SetJSON(map[string]string{"name": "gonet"}).
//		SetBearerToken("token").
//		Text()
//
// 支持 GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS, JSON/表单/原始/multipart 请求体,
// Basic/Bearer 认证, 请求头与查询参数, 上下文取消, 自动重试, 响应大小限制,
// 以及请求前/响应后钩子。下载文件用 Download, 上传大文件用 SetFileStream。
package client

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/laixhe/gonet/network/header"
)

// DefaultClient 默认客户端实例, baseURL 为空, 可直接使用或作为模板
var DefaultClient = NewClient("")

// Client HTTP 客户端, 链式构建请求
type Client struct {
	baseURL    string
	httpClient *http.Client

	// beforeRequest 请求发送前钩子, 按注册顺序执行
	beforeRequest []func(*http.Request) error
	// afterResponse 响应返回后钩子, 按注册顺序执行
	afterResponse []func(*http.Response) error

	// 请求级配置默认值, 请求可通过对应 Set 方法覆盖
	retryCount   int
	retryWait    time.Duration
	retryMaxWait time.Duration
	retryConds   []RetryCondition
	checkStatus  bool
	maxBodySize  int64
}

// NewClient 创建 HTTP 客户端, baseURL 为空时 URL 需为完整地址
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: DefaultHttpClient(),
	}
}

// NewClientWithHttpClient 使用自定义 http.Client 创建客户端
func NewClientWithHttpClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = DefaultHttpClient()
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Get 创建 GET 请求
func (c *Client) Get(URL string) *Request {
	return c.newRequest(http.MethodGet, URL)
}

// Post 创建 POST 请求
func (c *Client) Post(URL string) *Request {
	return c.newRequest(http.MethodPost, URL)
}

// Put 创建 PUT 请求
func (c *Client) Put(URL string) *Request {
	return c.newRequest(http.MethodPut, URL)
}

// Patch 创建 PATCH 请求
func (c *Client) Patch(URL string) *Request {
	return c.newRequest(http.MethodPatch, URL)
}

// Delete 创建 DELETE 请求
func (c *Client) Delete(URL string) *Request {
	return c.newRequest(http.MethodDelete, URL)
}

// Head 创建 HEAD 请求
func (c *Client) Head(URL string) *Request {
	return c.newRequest(http.MethodHead, URL)
}

// Options 创建 OPTIONS 请求
func (c *Client) Options(URL string) *Request {
	return c.newRequest(http.MethodOptions, URL)
}

// newRequest 创建请求
func (c *Client) newRequest(method, URL string) *Request {
	req := &Request{
		c:      c,
		Method: method,
		Header: make(http.Header),
	}
	u, err := url.Parse(joinURL(c.baseURL, URL))
	if err != nil {
		req.err = err
		return req
	}
	req.URL = u
	req.QueryParams = u.Query()
	return req
}

// joinURL 拼接 baseURL 与相对路径, 避免斜杠重复或缺失
func joinURL(base, ref string) string {
	if base == "" {
		return ref
	}
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(ref, "/")
}

// SetBaseURL 修改基础地址
func (c *Client) SetBaseURL(baseURL string) *Client {
	c.baseURL = baseURL
	return c
}

// SetTimeout 设置请求超时时间
func (c *Client) SetTimeout(d time.Duration) *Client {
	c.httpClient.Timeout = d
	return c
}

// SetHttpClient 替换内部 http.Client
func (c *Client) SetHttpClient(httpClient *http.Client) *Client {
	if httpClient != nil {
		c.httpClient = httpClient
	}
	return c
}

// SetTransport 设置传输层
func (c *Client) SetTransport(t http.RoundTripper) *Client {
	c.httpClient.Transport = t
	return c
}

// SetCheckRedirect 设置重定向策略
func (c *Client) SetCheckRedirect(fn func(req *http.Request, via []*http.Request) error) *Client {
	c.httpClient.CheckRedirect = fn
	return c
}

// SetCookieJar 设置 Cookie 容器
func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	c.httpClient.Jar = jar
	return c
}

// SetRetryCount 设置默认重试次数, 请求可通过 SetRetryCount 覆盖
func (c *Client) SetRetryCount(n int) *Client {
	c.retryCount = n
	return c
}

// SetRetryWaitTime 设置默认重试等待时间, 每次重试后翻倍
func (c *Client) SetRetryWaitTime(d time.Duration) *Client {
	c.retryWait = d
	return c
}

// SetRetryMaxWaitTime 设置默认最大重试等待时间
func (c *Client) SetRetryMaxWaitTime(d time.Duration) *Client {
	c.retryMaxWait = d
	return c
}

// AddRetryCondition 注册默认重试条件, 请求可通过 AddRetryCondition 追加
func (c *Client) AddRetryCondition(fn RetryCondition) *Client {
	if fn != nil {
		c.retryConds = append(c.retryConds, fn)
	}
	return c
}

// SetCheckStatus 设置默认非 2xx 状态码校验开关, 请求可通过 SetCheckStatus 覆盖
func (c *Client) SetCheckStatus(enable bool) *Client {
	c.checkStatus = enable
	return c
}

// SetMaxBodySize 设置默认响应体读取上限, 请求可通过 SetMaxBodySize 覆盖
func (c *Client) SetMaxBodySize(n int64) *Client {
	c.maxBodySize = n
	return c
}

// OnBeforeRequest 注册请求发送前钩子, 可修改请求或提前返回错误
func (c *Client) OnBeforeRequest(fn func(*http.Request) error) *Client {
	if fn != nil {
		c.beforeRequest = append(c.beforeRequest, fn)
	}
	return c
}

// OnAfterResponse 注册响应返回后钩子
func (c *Client) OnAfterResponse(fn func(*http.Response) error) *Client {
	if fn != nil {
		c.afterResponse = append(c.afterResponse, fn)
	}
	return c
}

// DefaultPooledTransport 返回带连接池与超时配置的默认 Transport
func DefaultPooledTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment, // 使用系统代理
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		DisableCompression:    false,
		MaxIdleConns:          200,              // 最大空闲连接数
		MaxIdleConnsPerHost:   100,              // 每主机最大空闲连接数
		MaxConnsPerHost:       100,              // 每主机最大连接数
		IdleConnTimeout:       30 * time.Second, // 空闲连接关闭时间
		TLSHandshakeTimeout:   2 * time.Second,  // TLS 握手超时
		ResponseHeaderTimeout: 2 * time.Second,  // 响应头超时
		ExpectContinueTimeout: 2 * time.Second,
	}
}

// DefaultHttpClient 返回使用默认连接池的 HTTP 客户端
func DefaultHttpClient() *http.Client {
	return &http.Client{
		Transport: DefaultPooledTransport(),
	}
}

// HttpRequest 创建带默认 User-Agent 的 HTTP 请求
func HttpRequest(method string, URL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, URL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(header.UserAgent, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")
	return req, nil
}

// CloseResponse 关闭响应体并复用连接
func CloseResponse(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
