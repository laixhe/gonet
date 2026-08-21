package oplatform

import (
	"errors"
	"time"
)

// Config 微信开放平台配置
type Config struct {
	// 唯一凭证 App ID
	AppID string `json:"app_id" mapstructure:"appid" toml:"appid" yaml:"appid"`
	// 密钥 App Secret
	Secret string `json:"secret" mapstructure:"secret" toml:"secret" yaml:"secret"`
	// Timeout 接口调用超时时间,0 表示使用默认 10 秒
	Timeout time.Duration `json:"timeout" mapstructure:"timeout" toml:"timeout" yaml:"timeout"`
}

// Check 检查
func (c *Config) Check() error {
	if c == nil {
		return errors.New("微信开放平台配置不能为空")
	}
	if c.AppID == "" {
		return errors.New("微信开放平台 appid 不能为空")
	}
	if c.Secret == "" {
		return errors.New("微信开放平台 secret 不能为空")
	}
	return nil
}
