package imagesearch

import (
	"errors"

	openApiClient "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	imageSearchClient "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/laixhe/gonet/protocol/gen/config/calibaba"
	"github.com/laixhe/gonet/xlog"
)

// 阿里云图像搜索

type SdkAliyunImageSearch struct {
	config *calibaba.ImageSearch
	client *imageSearchClient.Client
}

func (s *SdkAliyunImageSearch) Config() *calibaba.ImageSearch {
	return s.config
}

func (s *SdkAliyunImageSearch) Client() *imageSearchClient.Client {
	return s.client
}

func Init(config *calibaba.ImageSearch) (*SdkAliyunImageSearch, error) {
	if config == nil {
		return nil, errors.New("aliyun image search config is nil")
	}
	if config.AccessKeyId == "" {
		return nil, errors.New("aliyun image search config access_key_id is nil")
	}
	if config.AccessKeySecret == "" {
		return nil, errors.New("aliyun image search config access_key_secret is nil")
	}
	if config.Region == "" {
		return nil, errors.New("aliyun image search config region is nil")
	}
	if config.Endpoint == "" {
		return nil, errors.New("aliyun image search config endpoint is nil")
	}
	xlog.Debugf("aliyun image search config=%v", config)
	//
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
	// 设置实例名称
	request := new(imageSearchClient.DetailRequest).SetInstanceName(config.InstanceName)
	resp, err := client.Detail(request)
	if err != nil {
		return nil, err
	}
	if resp.Body == nil || resp.Body.Success == nil || !tea.BoolValue(resp.Body.Success) {
		return nil, errors.New("aliyun image search detail fail")
	}
	return &SdkAliyunImageSearch{
		config: config,
		client: client,
	}, nil
}
