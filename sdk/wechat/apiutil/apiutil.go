// Package apiutil 提供微信系列 SDK(miniprogram/oplatform/work)接口调用共用的请求封装与错误类型。
//
// 各 SDK 模块通过 ApiError 别名导出错误类型,例如:
//
//	var apiErr *miniprogram.ApiError
//	if errors.As(err, &apiErr) { _ = apiErr.ErrCode }
package apiutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"resty.dev/v3"
)

// 错误类别哨兵,配合 errors.Is 判断错误类型,例如:
//
//	if errors.Is(err, apiutil.ErrNetwork) { ... }
var (
	// ErrNetwork 网络错误(连接失败、超时、上下文取消等)
	ErrNetwork = errors.New("wechat api: network error")
	// ErrHTTP HTTP 非 2xx 响应
	ErrHTTP = errors.New("wechat api: http error")
	// ErrBusiness 业务错误(errcode != 0)
	ErrBusiness = errors.New("wechat api: business error")
	// ErrDecode 响应体解析失败
	ErrDecode = errors.New("wechat api: decode error")
)

// ErrKind 错误类别。
type ErrKind int

const (
	// ErrKindNetwork 网络错误
	ErrKindNetwork ErrKind = iota + 1
	// ErrKindHTTP HTTP 非 2xx 响应
	ErrKindHTTP
	// ErrKindBusiness 业务错误(errcode != 0)
	ErrKindBusiness
	// ErrKindDecode 响应体解析失败
	ErrKindDecode
)

// ApiError 微信接口调用错误。
//
// 通过 errors.As 提取错误详情,通过 errors.Is 判断错误类别:
//
//	var apiErr *miniprogram.ApiError
//	if errors.As(err, &apiErr) { _ = apiErr.ErrCode }
//	if errors.Is(err, apiutil.ErrNetwork) { ... }
type ApiError struct {
	Kind    ErrKind // 错误类别
	ErrCode int     // 微信错误码,仅 ErrKindBusiness 时有效
	ErrMsg  string  // 错误信息,请求失败时返回
	Status  int     // HTTP 状态码,仅 ErrKindHTTP 时有效
	kindErr error   // 对应的哨兵错误
}

// Unwrap 返回错误类别对应的哨兵错误,支持 errors.Is 判断。
func (e *ApiError) Unwrap() error {
	return e.kindErr
}

// newApiError 构造 ApiError 并关联哨兵错误。
func newApiError(kind ErrKind, errCode, status int, msg string) *ApiError {
	var sentinel error
	switch kind {
	case ErrKindNetwork:
		sentinel = ErrNetwork
	case ErrKindHTTP:
		sentinel = ErrHTTP
	case ErrKindBusiness:
		sentinel = ErrBusiness
	case ErrKindDecode:
		sentinel = ErrDecode
	}
	return &ApiError{Kind: kind, ErrCode: errCode, ErrMsg: msg, Status: status, kindErr: sentinel}
}

func (e *ApiError) Error() string {
	switch e.Kind {
	case ErrKindHTTP:
		return fmt.Sprintf("wechat api error: http %d, %s", e.Status, e.ErrMsg)
	case ErrKindBusiness:
		return fmt.Sprintf("wechat api error: errcode %d, %s", e.ErrCode, e.ErrMsg)
	default:
		return fmt.Sprintf("wechat api error: %s", e.ErrMsg)
	}
}

// FlexString 兼容 JSON 字符串与数字的字符串字段。
//
// 微信部分接口同一字段在不同场景下可能返回字符串或数字(如时间戳、性别),
// 直接使用 string 会导致整个响应体解析失败,使用本类型可同时兼容两者。
type FlexString string

// UnmarshalJSON 同时接受 JSON 字符串与数字,并忽略 null。
func (s *FlexString) UnmarshalJSON(data []byte) error {
	*s = ""
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = FlexString(str)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(data, &num); err == nil {
		*s = FlexString(num.String())
		return nil
	}
	return fmt.Errorf("apiutil: cannot unmarshal %s into FlexString", data)
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
		return nil, newApiError(ErrKindNetwork, 0, 0, err.Error())
	}
	data := httpResp.Bytes()
	if httpResp.IsStatusSuccess() {
		env := new(errEnvelope)
		if err := json.Unmarshal(data, env); err != nil {
			return nil, newApiError(ErrKindDecode, 0, httpResp.StatusCode(), httpResp.String())
		}
		if env.ErrCode != 0 {
			return nil, newApiError(ErrKindBusiness, env.ErrCode, 0, env.ErrMsg)
		}
		result := new(T)
		if err := json.Unmarshal(data, result); err != nil {
			return nil, newApiError(ErrKindDecode, 0, 0, err.Error())
		}
		return result, nil
	}
	return nil, newApiError(ErrKindHTTP, 0, httpResp.StatusCode(), httpResp.String())
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
		return nil, newApiError(ErrKindNetwork, 0, 0, err.Error())
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
			return nil, newApiError(ErrKindDecode, 0, httpResp.StatusCode(), httpResp.String())
		}
		return nil, newApiError(ErrKindBusiness, env.ErrCode, 0, env.ErrMsg)
	}
	return nil, newApiError(ErrKindHTTP, 0, httpResp.StatusCode(), httpResp.String())
}
