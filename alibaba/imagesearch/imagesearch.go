package imagesearch

import (
	"errors"
	"fmt"

	openApiClient "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	imageSearchClient "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

/*
imagesearch:
	access_key: access_key
	access_key_id: XXXXXX
	access_key_secret: XXXXXX
	region: cn-shenzhen
	endpoint: imagesearch.cn-shenzhen.aliyuncs.com
	instance_name: XXXXXX
*/

type Config struct {
	// 密钥ID
	AccessKeyId string `json:"access_key_id" mapstructure:"access_key_id" toml:"access_key_id" yaml:"access_key_id"`
	// 密钥
	AccessKeySecret string `json:"access_key_secret" mapstructure:"access_key_secret" toml:"access_key_secret" yaml:"access_key_secret"`
	// 访问类型
	AccessKey string `json:"access_key" mapstructure:"access_key" toml:"access_key" yaml:"access_key"`
	// 地域(如: cn-shenzhen )
	Region string `json:"region" mapstructure:"region" toml:"region" yaml:"region"`
	// 域名(如: imagesearch.cn-shenzhen.aliyuncs.com )
	Endpoint string `json:"endpoint" mapstructure:"endpoint" toml:"endpoint" yaml:"endpoint"`
	// 图像搜索实例名称(注意是实例名称不是实例ID)
	InstanceName string `json:"instance_name" mapstructure:"instance_name" toml:"instance_name" yaml:"instance_name"`
}

// Checking 检查
func (c *Config) Check() error {
	if c == nil {
		return errors.New("没有图像搜索配置")
	}
	if c.AccessKeyId == "" {
		return errors.New("没有图像搜索密钥ID配置")
	}
	if c.AccessKeySecret == "" {
		return errors.New("没有图像搜索密钥配置")
	}
	if c.AccessKey == "" {
		return errors.New("没有图像搜索访问类型配置")
	}
	if c.Region == "" {
		return errors.New("没有图像搜索地域配置")
	}
	if c.Endpoint == "" {
		return errors.New("没有图像搜索域名配置")
	}
	if c.InstanceName == "" {
		return errors.New("没有图像搜索实例名称配置")
	}
	return nil
}

type ImageSearchClient struct {
	config *Config
	client *imageSearchClient.Client
}

func Init(config *Config) (*ImageSearchClient, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	var cfg = new(openApiClient.Config).
		SetAccessKeyId(config.AccessKeyId).
		SetAccessKeySecret(config.AccessKeySecret).
		SetType(config.AccessKey).
		SetEndpoint(config.Endpoint).
		SetRegionId(config.Region)
	// 创建客户端
	client, err := imageSearchClient.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	request := new(imageSearchClient.DetailRequest).SetInstanceName(config.InstanceName)
	resp, err := client.Detail(request)
	if err != nil {
		return nil, err
	}
	if resp.Body == nil || resp.Body.Success == nil || !tea.BoolValue(resp.Body.Success) {
		return nil, fmt.Errorf("imagesearch detail fail: %d", tea.Int32Value(resp.StatusCode))
	}
	return &ImageSearchClient{
		config: config,
		client: client,
	}, nil
}
