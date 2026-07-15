package douyin

import "errors"

type Config struct {
	// 唯一凭证 App ID
	AppID string `json:"app_id" mapstructure:"app_id" toml:"app_id" yaml:"app_id"`
	// 密钥 App Secret
	AppSecret string `json:"app_secret" mapstructure:"app_secret" toml:"app_secret" yaml:"app_secret"`
	// 应用私钥(需要把 RSA 的私钥头部和尾部的标识去掉，并且整合成一行字符串)
	PrivateKey string `json:"private_key" mapstructure:"private_key" toml:"private_key" yaml:"private_key"`
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
