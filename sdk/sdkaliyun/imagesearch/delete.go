package imagesearch

import (
	"errors"

	imageSearchClient "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

func (s *SdkAliyunImageSearch) Delete(fileID string) error {
	request := new(imageSearchClient.DeleteImageRequest).
		SetInstanceName(s.config.InstanceName).
		SetProductId(fileID)
	// 调用 api
	resp, err := s.client.DeleteImage(request)
	if err != nil {
		return err
	}
	if resp.Body == nil || resp.Body.Success == nil || !tea.BoolValue(resp.Body.Success) || resp.Body.Code == nil || tea.Int32Value(resp.Body.Code) != 0 {
		if resp.Body.Message != nil && tea.StringValue(resp.Body.Message) != "" {
			return errors.New("aliyun image search delete fail: " + tea.StringValue(resp.Body.Message))
		}
		return errors.New("aliyun image search delete fail")
	}
	return nil
}
