package imagesearch

import (
	"fmt"

	imageSearchClient "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
	utilService "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// Add 添加图片
func (isc *ImageSearchClient) Add(req *imageSearchClient.AddImageAdvanceRequest, runtimeObjects ...*utilService.RuntimeOptions) error {
	/*
			 req := new(imageSearchClient.AddImageAdvanceRequest).
		        // 必填，图像搜索实例名称。注意是实例名称不是实例ID。购买后通过上云层管控台实例信息一栏查看：https://imagesearch.console.aliyun.com/overview
		        SetInstanceName("XXXXXXXX").
				// 必填，图片内容(io.Reader)，最多支持 4MB大小图片以及5s的传输等待时间。当前仅支持PNG、JPG、JPEG、BMP、GIF、WEBP、TIFF、PPM格式图片；
		        // 对于商品、商标、通用图片搜索，图片长和宽的像素必须都大于等于100且小于等于4096；
		        // 对于布料搜索，图片长和宽的像素必须都大于等于448且小于等于4096；
		        // 图像中不能带有旋转信息图片内容，最多支持 2MB大小图片以及5s的传输等待时间。当前仅支持jpg和png格式图片；
		        SetPicContentObject(b).
		        // 必填，图片名称，最多支持 256个字符。
		        // 1. ProductId + PicName唯一确定一张图片。
		        // 2. 如果多次添加图片具有相同的ProductId + PicName，以最后一次添加为准，前面添加的图片将被覆盖。
		        SetPicName("test").
		        // 必填，商品id，最多支持 256个字符。
		        // 一个商品可有多张图片。
		        SetProductId("test").
		        // 选填，图片类目。
		        // 1. 对于商品搜索：若设置类目，则以设置的为准；若不设置类目，将由系统进行类目预测，预测的类目结果可在Response中获取 。
		        // 2. 对于布料、商标、通用搜索：不论是否设置类目，系统会将类目设置为88888888。
		        SetCategoryId(2).
		        // 选填，是否需要进行主体识别，默认为true。
		        // 1.为true时，由系统进行主体识别，以识别的主体进行搜索，主体识别结果可在Response中获取。
		        // 2. 为false时，则不进行主体识别，以整张图进行搜索。
		        // 3.对于布料图片搜索，此参数会被忽略，系统会以整张图进行搜索。
		        SetCrop(true).
		        // 选填，图片的主体区域，格式为 x1,x2,y1,y2, 其中 x1,y1 是左上角的点，x2，y2是右下角的点。
		        // 设置的region 区域不要超过图片的边界。
		        // 若用户设置了Region，则不论Crop参数为何值，都将以用户输入Region进行搜索。
		        // 对于布料图片搜索，此参数会被忽略，系统会以整张图进行搜索。
		        SetRegion("167,477,220,407").
		        // 选填，用户自定义的内容，最多支持 4096个字符。
		        // 查询时会返回该字段。例如可添加图片的描述等文本。
		        SetCustomContent("this is a simple test!").
		        // 选填，整数类型属性，可用于查询时过滤，查询时会返回该字段。
		        // 例如不同的站点的图片/不同用户的图片，可以设置不同的IntAttr，查询时通过过滤来达到隔离的目的
		        SetIntAttr(100).
		        // 选填，字符串类型属性，最多支持 128个字符。可用于查询时过滤，查询时会返回该字段。
		        SetStrAttr("1")
	*/
	if req.InstanceName == nil {
		req.SetInstanceName(isc.config.InstanceName)
	}
	var runtimeObject = new(utilService.RuntimeOptions)
	if len(runtimeObjects) > 0 {
		runtimeObject = runtimeObjects[0]
	}
	resp, err := isc.client.AddImageAdvance(req, runtimeObject)
	if err != nil {
		return err
	}
	if resp.Body == nil || resp.Body.Success == nil || !tea.BoolValue(resp.Body.Success) || resp.Body.Code == nil || tea.Int32Value(resp.Body.Code) != 0 {
		if resp.Body.Message != nil && tea.StringValue(resp.Body.Message) != "" {
			return fmt.Errorf("imagesearch add fail: %d %s", tea.Int32Value(resp.Body.Code), tea.StringValue(resp.Body.Message))
		}
		return fmt.Errorf("imagesearch add fail: %d", tea.Int32Value(resp.StatusCode))
	}
	return nil
}
