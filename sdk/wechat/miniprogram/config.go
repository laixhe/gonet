package miniprogram

import (
	"errors"
	"time"
)

type Config struct {
	// 唯一凭证 App ID
	AppID string `json:"app_id" mapstructure:"appid" toml:"appid" yaml:"appid"`
	// 密钥 App Secret
	Secret string `json:"secret" mapstructure:"secret" toml:"secret" yaml:"secret"`
	// (可选)回调消息 Token
	Token string `json:"token" mapstructure:"token" toml:"token" yaml:"token"`
	// (可选)回调消息密钥 AESKey
	AesKey string `json:"aeskey" mapstructure:"aeskey" toml:"aeskey" yaml:"aeskey"`
	// Timeout 接口调用超时时间,0 表示使用默认 10 秒
	Timeout time.Duration `json:"timeout" mapstructure:"timeout" toml:"timeout" yaml:"timeout"`
}

func (c *Config) Check() error {
	if c == nil {
		return errors.New("微信小程序配置不能为空")
	}
	if c.AppID == "" {
		return errors.New("微信小程序 appid 不能为空")
	}
	if c.Secret == "" {
		return errors.New("微信小程序 secret 不能为空")
	}
	return nil
}
