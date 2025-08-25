package oplatform

import "errors"

// 微信开放平台配置
type Config struct {
	// 唯一凭证 App ID
	AppId string `json:"app_id" mapstructure:"appid" toml:"appid" yaml:"appid"`
	// 密钥 App Secret
	Secret string `json:"secret" mapstructure:"secret" toml:"secret" yaml:"secret"`
	// (可选)回调消息 Token
	Token string `json:"token" mapstructure:"token" toml:"token" yaml:"token"`
	// (可选)回调消息密钥 AESKey
	Aeskey string `json:"aeskey" mapstructure:"aeskey" toml:"aeskey" yaml:"aeskey"`
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
