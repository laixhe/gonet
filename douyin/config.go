package douyin

import "errors"

type Config struct {
	// 唯一凭证 App ID
	AppID string `json:"app_id" mapstructure:"appid" toml:"appid" yaml:"appid"`
	// 密钥 App Secret
	AppSecret string `json:"app_secret" mapstructure:"app_secret" toml:"app_secret" yaml:"app_secret"`
}

func (c *Config) Check() error {
	if c == nil {
		return errors.New("抖音开放平台配置不能为空")
	}
	if c.AppID == "" {
		return errors.New("抖音开放平台 app_id 不能为空")
	}
	if c.AppSecret == "" {
		return errors.New("抖音开放平台 app_secret 不能为空")
	}
	return nil
}
