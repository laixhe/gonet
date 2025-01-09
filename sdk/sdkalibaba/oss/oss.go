package oss

import (
	"errors"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"github.com/laixhe/gonet/protocol/gen/config/calibaba"
	"github.com/laixhe/gonet/xlog"
)

// 阿里云对象存储

type SdkOss struct {
	c      *calibaba.Oss
	client *ossv2.Client
}

var sdkOss *SdkOss

func Init(c *calibaba.Oss) error {
	if c == nil {
		return errors.New("oss config is nil")
	}
	if c.AccessKeyId == "" {
		return errors.New("oss config access_key_id is nil")
	}
	if c.AccessKeySecret == "" {
		return errors.New("oss config access_key_secret is nil")
	}
	if c.Region == "" {
		return errors.New("oss config region is nil")
	}
	if c.Endpoint == "" {
		return errors.New("oss config endpoint is nil")
	}
	if c.Bucket == "" {
		return errors.New("oss config bucket is nil")
	}
	xlog.Debugf("oss config=%v", c)
	//
	sdkOss = &SdkOss{
		c:      c,
		client: nil,
	}
	// doc https://help.aliyun.com/zh/oss
	cfg := ossv2.NewConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyId, c.AccessKeySecret)).
		WithRegion(c.Region).WithEndpoint(c.Endpoint)
	client := ossv2.NewClient(cfg)
	// 检查存储空间是否存在
	//isBucket, err := client.IsBucketExist(context.TODO(), c.Bucket)
	//if err != nil {
	//	return err
	//}
	//if !isBucket {
	//	return errors.New("oss bucket is not exist")
	//}
	sdkOss.client = client
	return nil
}
