package client

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/laixhe/gonet/network/header"
)

var DefaultClient = NewClient("")

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: DefaultHttpClient(),
	}
}

func (c *Client) Get(URL string) *Request {
	req := &Request{
		c:      c,
		Method: http.MethodGet,
	}
	reqURL, err := url.Parse(c.baseURL + URL)
	if err != nil {
		req.err = err
		return req
	}
	req.URL = reqURL
	req.QueryParams = reqURL.Query()
	return req
}

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

func DefaultHttpClient() *http.Client {
	return &http.Client{
		Transport: DefaultPooledTransport(),
	}
}

func HttpRequest(method string, URL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, URL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(header.UserAgent, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")
	return req, nil
}

func CloseResponse(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
