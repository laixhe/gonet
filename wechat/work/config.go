package work

import "errors"

// Config 企业微信配置
type Config struct {
	// 企业ID
	Corpid string `json:"corpid" mapstructure:"corpid" toml:"corpid" yaml:"corpid"`
	// 企业凭证密钥
	Corpsecret string `json:"corpsecret" mapstructure:"corpsecret" toml:"corpsecret" yaml:"corpsecret"`
	// 企业应用ID
	Agentid string `json:"agentid" mapstructure:"agentid" toml:"agentid" yaml:"agentid"`
}

// Check 检查
func (c *Config) Check() error {
	if c == nil {
		return errors.New("企业微信配置不能为空")
	}
	if c.Corpid == "" {
		return errors.New("企业微信 corpid 不能为空")
	}
	if c.Corpsecret == "" {
		return errors.New("企业微信 corpsecret 不能为空")
	}
	if c.Agentid == "" {
		return errors.New("企业微信 agentid 不能为空")
	}
	return nil
}
