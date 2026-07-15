package oplatform

import (
	"resty.dev/v3"

	"github.com/laixhe/gonet/sdk/wechat/oplatform/sns/oauth2"
)

// 开放平台

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
func (o *OpenPlatform) AccessToken(code string) (*oauth2.AccessTokenResponse, error) {
	return oauth2.AccessToken(o.httpClient, o.config.AppID, o.config.Secret, code)
}
