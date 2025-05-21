package imagesearch

import (
	"errors"
	"io"

	imageSearchClient "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
	utilService "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// Add 添加图片
func (s *SdkAliyunImageSearch) Add(body io.Reader, fileID, fileName string) error {
	var runtimeObject = new(utilService.RuntimeOptions)
	request := new(imageSearchClient.AddImageAdvanceRequest).
		SetInstanceName(s.config.InstanceName).
		SetProductId(fileID).
		SetPicName(fileName).
		SetPicContentObject(body)
	resp, err := s.client.AddImageAdvance(request, runtimeObject)
	if err != nil {
		return err
	}
	if resp.Body == nil || resp.Body.Success == nil || !tea.BoolValue(resp.Body.Success) || resp.Body.Code == nil || tea.Int32Value(resp.Body.Code) != 0 {
		if resp.Body.Message != nil && tea.StringValue(resp.Body.Message) != "" {
			return errors.New("aliyun image search add fail: " + tea.StringValue(resp.Body.Message))
		}
		return errors.New("aliyun image search add fail")
	}
	return nil
}
