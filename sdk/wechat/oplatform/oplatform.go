package oplatform

import (
	"context"
	"time"

	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/oplatform/internal/apiutil"
	"github.com/laixhe/gonet/sdk/wechat/oplatform/sns/oauth2"
)

// 开放平台

// defaultHTTPTimeout 微信接口默认超时时间。
const defaultHTTPTimeout = 10 * time.Second

// ApiError 微信接口调用错误,可通过 errors.As 获取错误码。
//
//	var apiErr *oplatform.ApiError
//	if errors.As(err, &apiErr) { _ = apiErr.ErrCode }
type ApiError = apiutil.ApiError

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
	httpClient.SetTimeout(defaultHTTPTimeout)
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
