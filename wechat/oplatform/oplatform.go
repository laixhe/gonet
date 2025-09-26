package oplatform

import (
	"resty.dev/v3"

	"github.com/laixhe/gonet/wechat/oplatform/sns/oauth2"
)

// Oplatform 开放平台
type Oplatform struct {
	config     *Config
	httpClient *resty.Client
}

func NewOplatform(config *Config) *Oplatform {
	if err := config.Check(); err != nil {
		panic(err)
	}
	httpClient := resty.New()
	httpClient.SetBaseURL("https://api.weixin.qq.com")
	return &Oplatform{
		config:     config,
		httpClient: httpClient,
	}
}

// Config 获取配置
func (o *Oplatform) Config() *Config {
	return o.config
}

// AccessToken 微信登录
// 通过 code 获取 access_token
func (o *Oplatform) AccessToken(code string) (*oauth2.AccessTokenResponse, error) {
	return oauth2.AccessToken(o.httpClient, o.config.AppId, o.config.Secret, code)
}
