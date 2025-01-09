package imagesearch

import (
	"errors"

	openApiClient "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	imageSearchClient "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/laixhe/gonet/protocol/gen/config/calibaba"
	"github.com/laixhe/gonet/xlog"
)

type SdkImageSearch struct {
	c      *calibaba.ImageSearch
	client *imageSearchClient.Client
}

var sdkImageSearch *SdkImageSearch

func Init(c *calibaba.ImageSearch) error {
	if c == nil {
		return errors.New("image search config is nil")
	}
	if c.AccessKeyId == "" {
		return errors.New("image search config access_key_id is nil")
	}
	if c.AccessKeySecret == "" {
		return errors.New("image search config access_key_secret is nil")
	}
	if c.Region == "" {
		return errors.New("image search config region is nil")
	}
	if c.Endpoint == "" {
		return errors.New("image search config endpoint is nil")
	}
	xlog.Debugf("image search config=%v", c)
	//
	sdkImageSearch = &SdkImageSearch{
		c:      c,
		client: nil,
	}
	//
	var cfg = new(openApiClient.Config).
		SetAccessKeyId(c.AccessKeyId).
		SetAccessKeySecret(c.AccessKeySecret).
		SetType(c.AccessKey).
		SetEndpoint(c.Endpoint).
		SetRegionId(c.Region)
	// 创建客户端
	client, err := imageSearchClient.NewClient(cfg)
	if err != nil {
		return err
	}
	//
	request := new(imageSearchClient.DetailRequest).SetInstanceName(c.InstanceName)
	resp, err := client.Detail(request)
	if err != nil {
		return err
	}
	if resp.Body == nil || resp.Body.Success == nil || !tea.BoolValue(resp.Body.Success) {
		return errors.New("image search detail fail")
	}
	sdkImageSearch.client = client
	return nil
}
