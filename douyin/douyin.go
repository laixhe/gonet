package douyin

import (
	credential "github.com/bytedance/douyin-openapi-credential-go/client"
	openApiSdkClient "github.com/bytedance/douyin-openapi-sdk-go/client"
)

type Douyin struct {
	config *Config
	client *openApiSdkClient.Client
}

func NewDouyin(config *Config) *Douyin {
	if err := config.Check(); err != nil {
		panic(err)
	}
	opt := new(credential.Config).
		SetClientKey(config.AppID).
		SetClientSecret(config.AppSecret)
	client, err := openApiSdkClient.NewClient(opt)
	if err != nil {
		panic(err)
	}
	return &Douyin{
		config: config,
		client: client,
	}
}

// Config 获取配置
func (d *Douyin) Config() *Config {
	return d.config
}
