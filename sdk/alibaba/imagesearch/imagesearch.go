package imagesearch

import (
	"errors"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	clientv4 "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
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
	AccessKeyID string `json:"access_key_id" mapstructure:"access_key_id" toml:"access_key_id" yaml:"access_key_id"`
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

func (c *Config) Check() error {
	if c == nil {
		return errors.New("没有图像搜索配置")
	}
	if c.AccessKeyID == "" {
		return errors.New("没有图像搜索密钥 access_key_id 配置")
	}
	if c.AccessKeySecret == "" {
		return errors.New("没有图像搜索密钥 access_key_secret 配置")
	}
	if c.AccessKey == "" {
		return errors.New("没有图像搜索访问类型 access_key 配置")
	}
	if c.Region == "" {
		return errors.New("没有图像搜索地域 region 配置")
	}
	if c.Endpoint == "" {
		return errors.New("没有图像搜索域名 endpoint 配置")
	}
	if c.InstanceName == "" {
		return errors.New("没有图像搜索实例名称 instance_name 配置")
	}
	return nil
}

type ISClient struct {
	config *Config
	client *clientv4.Client
}

func Init(config *Config) (*ISClient, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	var cfg = new(openapi.Config).
		SetAccessKeyId(config.AccessKeyID).
		SetAccessKeySecret(config.AccessKeySecret).
		SetType(config.AccessKey).
		SetEndpoint(config.Endpoint).
		SetRegionId(config.Region)
	// 创建客户端
	client, err := clientv4.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	request := new(clientv4.DetailRequest).SetInstanceName(config.InstanceName)
	resp, err := client.Detail(request)
	if err != nil {
		return nil, err
	}
	if resp.Body == nil || resp.Body.Success == nil || !tea.BoolValue(resp.Body.Success) {
		return nil, fmt.Errorf("imagesearch detail fail: %d", tea.Int32Value(resp.StatusCode))
	}
	return &ISClient{
		config: config,
		client: client,
	}, nil
}
