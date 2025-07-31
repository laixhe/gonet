package oss

import (
	"errors"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

/*
alibaba_oss:
  access_key_id: XXXXXX
  access_key_secret: XXXXXX
  region: cn-shenzhen
  endpoint: https://oss-cn-shenzhen.aliyuncs.com
  bucket: test
*/

// Config 对象存储配置
type Config struct {
	// 标识用户ID
	AccessKeyId string `json:"access_key_id" mapstructure:"access_key_id" toml:"access_key_id" yaml:"access_key_id"`
	// 密钥
	AccessKeySecret string `json:"access_key_secret" mapstructure:"access_key_secret" toml:"access_key_secret" yaml:"access_key_secret"`
	// 地域(如: cn-shenzhen)
	Region string `json:"region" mapstructure:"region" toml:"region" yaml:"region"`
	// 访问域名(如: https://oss-cn-shenzhen.aliyuncs.com)
	Endpoint string `json:"endpoint" mapstructure:"endpoint" toml:"endpoint" yaml:"endpoint"`
	// 桶名(存储空间如: test)
	Bucket string `json:"bucket" mapstructure:"bucket" toml:"bucket" yaml:"bucket"`
}

// Checking 检查
func (c *Config) Check() error {
	if c == nil {
		return errors.New("没有对象存储配置")
	}
	if c.AccessKeyId == "" {
		return errors.New("没有对象存储访问密钥ID配置")
	}
	if c.AccessKeySecret == "" {
		return errors.New("没有对象存储访问密钥配置")
	}
	if c.Endpoint == "" {
		return errors.New("没有对象存储访问域名配置")
	}
	if c.Bucket == "" {
		return errors.New("没有对象存储桶名配置")
	}
	return nil
}

type OssClient struct {
	config *Config
	client *ossv2.Client
}

func Init(config *Config) (*OssClient, error) {
	if err := config.Check(); err != nil {
		return nil, err
	}
	// doc https://help.aliyun.com/zh/oss
	cfg := ossv2.NewConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKeyId, config.AccessKeySecret)).
		WithRegion(config.Region).WithEndpoint(config.Endpoint)
	client := ossv2.NewClient(cfg)
	// 检查存储空间是否存在
	//isBucket, err := client.IsBucketExist(context.TODO(), config.Bucket)
	//if err != nil {
	//	return nil, err
	//}
	//if !isBucket {
	//	return nil, errors.New("oss bucket is not exist")
	//}
	return &OssClient{config: config, client: client}, nil
}
