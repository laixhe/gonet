package sdkoss

import (
	"errors"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/laixhe/gonet/xlog"

	"github.com/laixhe/gonet/protocol/gen/config/coss"
)

// 阿里云对象存储

type SdkOss struct {
	c      *coss.Oss
	client *oss.Client
}

var sdkOss *SdkOss

func Init(c *coss.Oss) error {
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
	cfg := oss.NewConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyId, c.AccessKeySecret)).
		WithRegion(c.Region).WithEndpoint(c.Endpoint)
	client := oss.NewClient(cfg)
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
