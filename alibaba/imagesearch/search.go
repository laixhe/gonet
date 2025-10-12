package imagesearch

import (
	"fmt"

	imageSearchClient "github.com/alibabacloud-go/imagesearch-20201214/v4/client"
	utilService "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

func (isc *ImageSearchClient) SearchByPic(req *imageSearchClient.SearchImageByPicAdvanceRequest, runtimeObjects ...*utilService.RuntimeOptions) ([]*imageSearchClient.SearchImageByPicResponseBodyAuctions, error) {
	/*
			req := new(imageSearchClient.SearchImageByPicAdvanceRequest).
		        // 必填，图像搜索实例名称。注意是实例名称不是实例ID。购买后通过上云层管控台实例信息一栏查看：https://imagesearch.console.aliyun.com/overview
		        SetInstanceName("xxxxxxxxxx").
		        // 必填，图片内容(io.Reader)，最多支持 4MB大小图片以及5s的传输等待时间。当前仅支持PNG、JPG、JPEG、BMP、GIF、WEBP、TIFF、PPM格式图片；
		        // 对于商品、商标、通用图片搜索，图片长和宽的像素必须都大于等于100且小于等于4096；
		        // 对于布料搜索，图片长和宽的像素必须都大于等于448且小于等于4096；
		        // 图像中不能带有旋转信息
		        SetPicContentObject(b).
		        // 选填，商品类目。
		        // 1. 对于商品搜索：若设置类目，则以设置的为准；若不设置类目，将由系统进行类目预测，预测的类目结果可在Response中获取 。
		        // 2. 对于布料、商标、通用搜索：不论是否设置类目，系统会将类目设置为88888888。
		        SetCategoryId(2).
		        // 选填，返回结果的数目。取值范围：1-100。默认值：10。
		        SetNum(10).
		        // 选填，返回结果的起始位置。取值范围：0-499。默认值：0
		        SetStart(0).
		        // 选填，过滤条件
		        // int_attr支持的操作符有>、>=、<、<=、=，str_attr支持的操作符有=和!=，多个条件之支持AND和OR进行连接。
		        // 示例：
		        //  1. 根据IntAttr过滤结果，int_attr>=100
		        //  2. 根据StrAttr过滤结果，str_attr!="value1"
		        //  3. 根据IntAttr和StrAttr联合过滤结果，int_attr=1000 AND str_attr="value1"
		        SetFilter("int_attr=101 OR str_attr=\"2\"").
		        // 选填，是否需要进行主体识别，默认为true。
		        // 1.为true时，由系统进行主体识别，以识别的主体进行搜索，主体识别结果可在Response中获取。
		        // 2. 为false时，则不进行主体识别，以整张图进行搜索。
		        // 3.对于布料图片搜索，此参数会被忽略，系统会以整张图进行搜索。
		        SetCrop(true).
		        // 选填，图片的主体区域，格式为 x1,x2,y1,y2, 其中 x1,y1 是左上角的点，x2，y2是右下角的点。
		        // 设置的region 区域不要超过图片的边界。
		        // 若用户设置了Region，则不论Crop参数为何值，都将以用户输入Region进行搜索。
		        // 3.对于布料图片搜索，此参数会被忽略，系统会以整张图进行搜索。
		        SetRegion("167,476,220,407").
		        // 选填,若为true则响应数据根据ProductId进行返回。
		        SetDistinctProductId(true)
	*/
	var runtimeObject = new(utilService.RuntimeOptions)
	if len(runtimeObjects) > 0 {
		runtimeObject = runtimeObjects[0]
	}
	if req.InstanceName == nil {
		req.SetInstanceName(isc.config.InstanceName)
	}
	resp, err := isc.client.SearchImageByPicAdvance(req, runtimeObject)
	if err != nil {
		return nil, err
	}
	if resp.Body == nil || resp.Body.Success == nil || !tea.BoolValue(resp.Body.Success) || resp.Body.Code == nil || tea.Int32Value(resp.Body.Code) != 0 {
		if resp.Body.Msg != nil && tea.StringValue(resp.Body.Msg) != "" {
			return nil, fmt.Errorf("imagesearch search pic fail: %d %s", tea.Int32Value(resp.Body.Code), tea.StringValue(resp.Body.Msg))
		}
		return nil, fmt.Errorf("imagesearch search pic fail: %d", tea.Int32Value(resp.StatusCode))
	}
	return resp.Body.Auctions, nil
}
