package imagesearch

import (
	"fmt"

	imageSearchClient "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

func (isc *ImageSearchClient) Delete(req *imageSearchClient.DeleteImageRequest) error {
	/*
			req := new(imageSearchClient.DeleteImageRequest).
		        // 必填，图像搜索实例名称。注意是实例名称不是实例ID。购买后通过上云层管控台实例信息一栏查看：https://imagesearch.console.aliyun.com/overview
		        SetInstanceName("XXXXXXXX").
		        // 必填，图片名称，最多支持 256个字符。
		        // 1. ProductId + PicName唯一确定一张图片。
		        SetPicName("test").
		        // 选填，图片名称。若不指定本参数，则删除ProductId下所有图片；若指定本参数，则删除ProductId+PicName指定的图片。
		        SetProductId("php").
		        // 选填,若为true则根据filter进行删除。
		        SetIsDeleteByFilter(false).
		        SetFilter("intattr3=xxx")
	*/
	if req.InstanceName == nil {
		req.SetInstanceName(isc.config.InstanceName)
	}
	resp, err := isc.client.DeleteImage(req)
	if err != nil {
		return err
	}
	if resp.Body == nil || resp.Body.Success == nil || !tea.BoolValue(resp.Body.Success) || resp.Body.Code == nil || tea.Int32Value(resp.Body.Code) != 0 {
		if resp.Body.Message != nil && tea.StringValue(resp.Body.Message) != "" {
			return fmt.Errorf("imagesearch delete fail: %d %s", tea.Int32Value(resp.Body.Code), tea.StringValue(resp.Body.Message))
		}
		return fmt.Errorf("imagesearch delete fail: %d", tea.Int32Value(resp.StatusCode))
	}
	return nil
}
