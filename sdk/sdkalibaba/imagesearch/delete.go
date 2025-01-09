package imagesearch

import (
	"errors"

	imageSearchClient "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

func Delete(fileID string) error {
	request := new(imageSearchClient.DeleteImageRequest).
		SetInstanceName(sdkImageSearch.c.InstanceName).
		SetProductId(fileID)
	// 调用 api
	resp, err := sdkImageSearch.client.DeleteImage(request)
	if err != nil {
		return err
	}
	if resp.Body == nil || resp.Body.Success == nil || !tea.BoolValue(resp.Body.Success) || resp.Body.Code == nil || tea.Int32Value(resp.Body.Code) != 0 {
		if resp.Body.Message != nil && tea.StringValue(resp.Body.Message) != "" {
			return errors.New("image search delete fail: " + tea.StringValue(resp.Body.Message))
		}
		return errors.New("image search delete fail")
	}
	return nil
}
