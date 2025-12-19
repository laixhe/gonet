package oplatform

import "errors"

// Config 微信开放平台配置
type Config struct {
	// 唯一凭证 App ID
	AppId string `json:"app_id" mapstructure:"appid" toml:"appid" yaml:"appid"`
	// 密钥 App Secret
	Secret string `json:"secret" mapstructure:"secret" toml:"secret" yaml:"secret"`
}

// Check 检查
func (c *Config) Check() error {
	if c == nil {
		return errors.New("微信开放平台配置不能为空")
	}
	if c.AppId == "" {
		return errors.New("微信开放平台 AppId 不能为空")
	}
	if c.Secret == "" {
		return errors.New("微信开放平台 Secret 不能为空")
	}
	return nil
}
