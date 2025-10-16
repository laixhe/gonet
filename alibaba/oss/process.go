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
// objectName 图片
// watermarkObjectName 水印图片
// watermarkWidth 水印宽度
// acls 访问权限
func (oc *OssClient) ProcessImageWatermark(ctx context.Context, objectName string, watermarkObjectName string, watermarkWidth int, acls ...ossv2.ObjectACLType) (string, error) {
	// 处理水印
	watermarkProcess := watermarkObjectName
	if watermarkWidth > 0 {
		watermarkProcess += fmt.Sprintf("?x-oss-process=image/resize,w_%d", watermarkWidth)
	}
	watermarkBase64 := base64.RawURLEncoding.EncodeToString([]byte(watermarkProcess))
	// 存储位置
	targetImageTime := time.Now().Format("20060102150405")
	targetImageName := objectName + "_watermark" + targetImageTime
	splitObjectName := strings.Split(objectName, ".")
	if len(splitObjectName) >= 2 {
		index := len(splitObjectName) - 2
		splitObjectName[index] = fmt.Sprintf("%s_watermark%s", splitObjectName[index], targetImageTime)
		targetImageName = strings.Join(splitObjectName, ".")
	}
	targetBase64 := base64.URLEncoding.EncodeToString([]byte(targetImageName))
	// 指定处理指令
	process := fmt.Sprintf("image/auto-orient,1/watermark,image_%s,t_60,g_se,x_20,y_20", watermarkBase64)                      // 水印
	process += fmt.Sprintf("|sys/saveas,o_%s,b_%s", targetBase64, base64.URLEncoding.EncodeToString([]byte(oc.config.Bucket))) // 存储
	request := &ossv2.ProcessObjectRequest{
		Bucket:  ossv2.Ptr(oc.config.Bucket),
		Key:     ossv2.Ptr(objectName),
		Process: ossv2.Ptr(process),
	}
	_, err := oc.client.ProcessObject(ctx, request)
	if err != nil {
		return "", err
	}
	if len(acls) > 0 {
		_ = oc.SetObjectACL(ctx, targetImageName)
	}
	return targetImageName, nil
}
