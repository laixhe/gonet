package oss

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// ProcessImageWatermark 添加水印图片
// imageObjectName 原图片
// watermarkObjectName 水印图片
// saveObjectName 保存图片
// watermarkG 水印在图片中的位置 nw=左上 north=中上 ne=右上 west=左中 center=中部 east=右中 sw=左下 south=中下 se=右下
// watermarkP 水印按要添加水印的原图，来进行百分比缩放 1~100
// watermarkT 水印透明度百分比 0~100
// watermarkY 水印右下角 Y 位置 0~4096
// watermarkX 水印角 X 位置 0~4096
// acls 访问权限
func (oc *OssClient) ProcessImageWatermark(ctx context.Context,
	imageObjectName string, watermarkObjectName string, saveObjectName string,
	watermarkG string, watermarkP int, watermarkT int, watermarkY int, watermarkX int,
	acls ...ossv2.ObjectACLType) (string, error) {
	// 检查参数
	switch watermarkG {
	case "nw":
	case "north":
	case "ne":
	case "west":
	case "center":
	case "east":
	case "sw":
	case "south":
	case "se":
	default:
		watermarkG = "se"
	}
	watermarkP = max(watermarkP, 1)
	watermarkP = min(watermarkP, 100)
	watermarkT = max(watermarkT, 0)
	watermarkT = min(watermarkT, 100)
	watermarkY = max(watermarkY, 0)
	watermarkY = min(watermarkY, 4096)
	watermarkX = max(watermarkX, 0)
	watermarkX = min(watermarkX, 4096)
	// 处理水印
	watermarkProcess := watermarkObjectName
	if watermarkP > 0 {
		watermarkProcess += fmt.Sprintf("?x-oss-process=image/resize,P_%d", watermarkP)
	}
	watermarkBase64 := base64.RawURLEncoding.EncodeToString([]byte(watermarkProcess))
	// 存储位置
	if saveObjectName == "" {
		targetImageTime := time.Now().Format("20060102150405")
		targetImageName := imageObjectName + "_watermark" + targetImageTime
		splitObjectName := strings.Split(imageObjectName, ".")
		if len(splitObjectName) >= 2 {
			index := len(splitObjectName) - 2
			splitObjectName[index] = fmt.Sprintf("%s_watermark%s",
				splitObjectName[index], targetImageTime)
			targetImageName = strings.Join(splitObjectName, ".")
		}
		saveObjectName = targetImageName
	}
	targetBase64 := base64.URLEncoding.EncodeToString([]byte(saveObjectName))
	// 指定处理指令
	process := fmt.Sprintf("image/auto-orient,1/watermark,image_%s,g_%s,t_%d,y_%d,x_%d",
		watermarkBase64, watermarkG, watermarkT, watermarkY, watermarkX) // 水印
	process += fmt.Sprintf("|sys/saveas,o_%s,b_%s",
		targetBase64, base64.URLEncoding.EncodeToString([]byte(oc.config.Bucket))) // 存储
	request := &ossv2.ProcessObjectRequest{
		Bucket:  ossv2.Ptr(oc.config.Bucket),
		Key:     ossv2.Ptr(imageObjectName),
		Process: ossv2.Ptr(process),
	}
	_, err := oc.client.ProcessObject(ctx, request)
	if err != nil {
		return "", err
	}
	if len(acls) > 0 {
		_ = oc.SetObjectACL(ctx, saveObjectName, acls[0])
	}
	return saveObjectName, nil
}
