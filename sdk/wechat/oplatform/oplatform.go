package oplatform

import (
	"context"
	"time"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/apiutil"
	"github.com/laixhe/gonet/sdk/wechat/oplatform/sns/oauth2"
)

// 开放平台

// defaultHTTPTimeout 微信接口默认超时时间。
const defaultHTTPTimeout = 10 * time.Second

// ApiError 微信接口调用错误,可通过 errors.As 获取错误码,通过 errors.Is 判断错误类别。
//
//	var apiErr *oplatform.ApiError
//	if errors.As(err, &apiErr) { _ = apiErr.ErrCode }
//	if errors.Is(err, ErrNetwork) { ... }
type ApiError = apiutil.ApiError

// 错误类别哨兵,可通过 errors.Is 判断错误类型。
var (
	ErrNetwork  = apiutil.ErrNetwork  // 网络错误(连接失败、超时、上下文取消)
	ErrHTTP     = apiutil.ErrHTTP     // HTTP 非 2xx 响应
	ErrBusiness = apiutil.ErrBusiness // 业务错误(errcode != 0)
	ErrDecode   = apiutil.ErrDecode   // 响应体解析失败
)

type OpenPlatform struct {
	config     *Config
	httpClient *resty.Client
}

func NewOpenPlatform(config *Config) *OpenPlatform {
	if err := config.Check(); err != nil {
		panic(err)
	}
	httpClient := resty.New()
	httpClient.SetBaseURL("https://api.weixin.qq.com")
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}
	httpClient.SetTimeout(timeout)
	return &OpenPlatform{
		config:     config,
		httpClient: httpClient,
	}
}

func (o *OpenPlatform) Config() *Config {
	return o.config
}

// AccessToken 微信登录
// 通过 code 获取 access_token
func (o *OpenPlatform) AccessToken(ctx context.Context, code string) (*oauth2.AccessTokenResponse, error) {
	return oauth2.AccessToken(ctx, o.httpClient, o.config.AppID, o.config.Secret, code)
}
