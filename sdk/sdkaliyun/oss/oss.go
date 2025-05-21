package oss

import (
	"errors"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"github.com/laixhe/gonet/protocol/gen/config/calibaba"
	"github.com/laixhe/gonet/xlog"
)

// 阿里云对象存储

type SdkAliyunOss struct {
	config *calibaba.Oss
	client *ossv2.Client
}

func (s *SdkAliyunOss) Config() *calibaba.Oss {
	return s.config
}

func (s *SdkAliyunOss) Client() *ossv2.Client {
	return s.client
}

func Init(config *calibaba.Oss) (*SdkAliyunOss, error) {
	if config == nil {
		return nil, errors.New("aliyun oss config is nil")
	}
	if config.AccessKeyId == "" {
		return nil, errors.New("aliyun oss config access_key_id is nil")
	}
	if config.AccessKeySecret == "" {
		return nil, errors.New("aliyun oss config access_key_secret is nil")
	}
	if config.Region == "" {
		return nil, errors.New("aliyun oss config region is nil")
	}
	if config.Endpoint == "" {
		return nil, errors.New("aliyun oss config endpoint is nil")
	}
	if config.Bucket == "" {
		return nil, errors.New("aliyun oss config bucket is nil")
	}
	xlog.Debugf("aliyun oss config=%v", config)
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
	//	return nil, errors.New("aliyun oss bucket is not exist")
	//}
	return &SdkAliyunOss{
		config: config,
		client: client,
	}, nil
}
