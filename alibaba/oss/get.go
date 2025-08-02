package oss

import (
	"context"
	"io"

	ossv2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// Get 简单下载
func (oc *OssClient) Get(ctx context.Context, objectName string) ([]byte, error) {
	// 创建获取对象的请求
	request := &ossv2.GetObjectRequest{
		Bucket: ossv2.Ptr(oc.config.Bucket), // 存储空间名称
		Key:    ossv2.Ptr(objectName),       // 对象名称
	}
	// 执行获取对象的操作并处理结果
	result, err := oc.client.GetObject(ctx, request)
	if err != nil {
		return nil, err
	}
	// 确保在函数结束时关闭响应体
	defer result.Body.Close()
	return io.ReadAll(result.Body)
}
