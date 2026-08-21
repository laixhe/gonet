package client

import (
	"io"
	"net/url"

	"github.com/laixhe/gonet/network/header"
)

// Request HTTP 请求, 通过 Client.Get 等创建后链式构建
type Request struct {
	c           *Client
	err         error
	Method      string
	URL         *url.URL
	QueryParams url.Values
	Result      any
}

// SetQueryParam 设置单个查询参数
func (r *Request) SetQueryParam(param, value string) *Request {
	r.QueryParams.Set(param, value)
	return r
}

// SetQueryParams 批量设置查询参数
func (r *Request) SetQueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.SetQueryParam(k, v)
	}
	return r
}

// Text 执行请求并返回响应文本
func (r *Request) Text() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	r.URL.RawQuery = r.QueryParams.Encode()
	req, err := HttpRequest(r.Method, r.URL.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set(header.Accept, "*/*")

	resp, err := r.c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer CloseResponse(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
