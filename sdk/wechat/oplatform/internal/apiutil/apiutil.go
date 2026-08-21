// Package apiutil 提供微信开放平台接口调用共用的请求封装与错误类型。
package apiutil

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"resty.dev/v3"
)

// ApiError 微信接口调用错误。
//
// 可通过 errors.As 从返回值中提取,例如:
//
//	var apiErr *oplatform.ApiError
//	if errors.As(err, &apiErr) { _ = apiErr.ErrCode }
type ApiError struct {
	ErrCode int    // 微信错误码,0 表示成功
	ErrMsg  string // 错误信息,请求失败时返回
	Status  int    // HTTP 状态码,0 表示非 HTTP 错误(如网络错误)
}

func (e *ApiError) Error() string {
	if e.Status > 0 {
		return fmt.Sprintf("wechat api error: http %d, errcode %d, %s", e.Status, e.ErrCode, e.ErrMsg)
	}
	return fmt.Sprintf("wechat api error: errcode %d, %s", e.ErrCode, e.ErrMsg)
}

// errEnvelope 用于从响应体中提取 errcode/errmsg。
type errEnvelope struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// newRequest 构造带公共设置的请求。
func newRequest(ctx context.Context, client *resty.Client, path string, query map[string]string, body any) *resty.Request {
	req := client.R().SetContext(ctx)
	if len(query) > 0 {
		req.SetQueryParams(query)
	}
	if body != nil {
		req.SetBody(body)
	}
	// 原生 encoding/json 会对 & < > 做 HTML 转义,微信侧部分参数会因此解析错误,统一禁用
	req.SetJSONEscapeHTML(false)
	return req
}

// Get 发起 GET 请求并解析 JSON 响应。
func Get[T any](ctx context.Context, client *resty.Client, path string, query map[string]string) (*T, error) {
	return doJSON[T](ctx, client, "GET", path, query, nil)
}

// Post 发起 POST 请求并解析 JSON 响应。
func Post[T any](ctx context.Context, client *resty.Client, path string, query map[string]string, body any) (*T, error) {
	return doJSON[T](ctx, client, "POST", path, query, body)
}

// doJSON 发起请求,统一处理网络错误、非 2xx 与业务错误(errcode != 0)。
func doJSON[T any](ctx context.Context, client *resty.Client, method, path string, query map[string]string, body any) (*T, error) {
	httpResp, err := newRequest(ctx, client, path, query, body).Execute(method, path)
	if err != nil {
		return nil, &ApiError{ErrCode: -1, ErrMsg: err.Error()}
	}
	data := httpResp.Bytes()
	if httpResp.IsStatusSuccess() {
		env := new(errEnvelope)
		if err := json.Unmarshal(data, env); err != nil {
			return nil, &ApiError{ErrCode: httpResp.StatusCode(), Status: httpResp.StatusCode(), ErrMsg: httpResp.String()}
		}
		if env.ErrCode != 0 {
			return nil, &ApiError{ErrCode: env.ErrCode, ErrMsg: env.ErrMsg}
		}
		result := new(T)
		if err := json.Unmarshal(data, result); err != nil {
			return nil, &ApiError{ErrCode: 500, ErrMsg: err.Error()}
		}
		return result, nil
	}
	return nil, &ApiError{ErrCode: httpResp.StatusCode(), Status: httpResp.StatusCode(), ErrMsg: httpResp.String()}
}

// BinaryResponse 二进制响应(如图片)。
type BinaryResponse struct {
	ContentType string // 响应 Content-Type
	Data        []byte // 响应体
}

// PostBinary 发起 POST 请求,成功时返回原始二进制(如图片),失败时解析 JSON 错误。
func PostBinary(ctx context.Context, client *resty.Client, path string, query map[string]string, body any) (*BinaryResponse, error) {
	httpResp, err := newRequest(ctx, client, path, query, body).Execute("POST", path)
	if err != nil {
		return nil, &ApiError{ErrCode: -1, ErrMsg: err.Error()}
	}
	data := httpResp.Bytes()
	if httpResp.IsStatusSuccess() {
		contentType := httpResp.Header().Get("Content-Type")
		if strings.HasPrefix(contentType, "image") {
			return &BinaryResponse{ContentType: contentType, Data: data}, nil
		}
		// 非图片响应为 JSON 错误
		env := new(errEnvelope)
		if err := json.Unmarshal(data, env); err != nil {
			return nil, &ApiError{ErrCode: httpResp.StatusCode(), Status: httpResp.StatusCode(), ErrMsg: httpResp.String()}
		}
		return nil, &ApiError{ErrCode: env.ErrCode, ErrMsg: env.ErrMsg}
	}
	return nil, &ApiError{ErrCode: httpResp.StatusCode(), Status: httpResp.StatusCode(), ErrMsg: httpResp.String()}
}
