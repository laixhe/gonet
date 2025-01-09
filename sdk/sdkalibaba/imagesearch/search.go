package imagesearch

import (
	"errors"
	"io"

	imageSearchClient "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
	utilService "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

func SearchImage(body io.Reader) ([]*imageSearchClient.SearchImageByPicResponseBodyAuctions, error) {
	var runtimeObject = new(utilService.RuntimeOptions)
	request := new(imageSearchClient.SearchImageByPicAdvanceRequest).
		SetInstanceName(sdkImageSearch.c.InstanceName).
		SetPicContentObject(body)
	resp, err := sdkImageSearch.client.SearchImageByPicAdvance(request, runtimeObject)
	if err != nil {
		return nil, err
	}
	if resp.Body == nil || resp.Body.Success == nil || !tea.BoolValue(resp.Body.Success) || resp.Body.Code == nil || tea.Int32Value(resp.Body.Code) != 0 {
		if resp.Body.Msg != nil && tea.StringValue(resp.Body.Msg) != "" {
			return nil, errors.New("image search pic fail: " + tea.StringValue(resp.Body.Msg))
		}
		return nil, errors.New("image search pic fail")
	}
	return resp.Body.Auctions, nil
}
